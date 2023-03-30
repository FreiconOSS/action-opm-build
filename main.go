package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/antchfx/xmlquery"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gitlab.itsm.freicon.de/otrs/tools/opmbuilder/internal"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "opmbuilder",
		Short: "opmbuiler builds opm",
		Long:  ``,
	}

	buildCmd := &cobra.Command{
		Use:   "build",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			version, _ := cmd.Flags().GetString("version")
			output, _ := cmd.Flags().GetString("output")

			opm, err := buildOpm(args[0], version)
			if err != nil {
				return err
			}
			f, err := os.OpenFile(output, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 777)
			if err != nil {
				return err
			}
			_, err = f.Write(opm)
			if err1 := f.Close(); err == nil {
				err = err1
			}
			return err
		},
	}
	buildCmd.Flags().String("version", "", "version")
	buildCmd.Flags().String("output", "", "output")

	uploadCmd := &cobra.Command{
		Use:   "upload",
		Args:  cobra.ExactArgs(2),
		Short: "",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := PostFile(fmt.Sprintf("%s%s", "https://opmly.itsm.freicon.de/upload/", args[0]), args[1])
			if err != nil {
				return err
			}
			return nil
		},
	}

	unpackCmd := newUnpackCmd()

	rootCmd.AddCommand(uploadCmd, buildCmd, unpackCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return
}

func newUnpackCmd() *cobra.Command {
	unpackCmd := &cobra.Command{
		Use:   "unpack",
		Args:  cobra.ExactArgs(1),
		Short: "",
		Long:  ``,
		RunE: func(cmd *cobra.Command, args []string) error {
			target, _ := cmd.Flags().GetString("target")

			content, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}
			doc, err := xmlquery.Parse(bytes.NewBuffer(content))
			if err != nil {
				return err
			}
			root := xmlquery.FindOne(doc, "//otrs_package")
			if root == nil {
				root = xmlquery.FindOne(doc, "//otobo_package")
			}
			name := root.SelectElement("Name").FirstChild.Data

			if len(target) == 0 {
				target = name
			}

			if !strings.HasSuffix(target, "/") {
				target = target + "/"
			}

			n := root.SelectElement("//Filelist")
			for _, node := range n.SelectElements("File") {
				parts := strings.Split(node.SelectAttr("Location"), "/")
				path := target + strings.Join(parts[:len(parts)-1], "/")
				if _, err := os.Stat(path); os.IsNotExist(err) {
					err = os.MkdirAll(path, 0777)
					if err != nil {
						return err
					}
				}

				str, err := base64.StdEncoding.DecodeString(node.FirstChild.Data)
				if err != nil {
					return err
				}
				err = os.WriteFile(target+node.SelectAttr("Location"), []byte(str), 0644)
				if err != nil {
					return err
				}

				// remove file and encoding
				node.FirstChild.Data = ""
				for i, attr := range node.Attr {
					if attr.Name.Local == "Encode" {
						node.Attr = append(node.Attr[:i], node.Attr[i+1:]...)
					}
				}
			}

			err = os.WriteFile(target+"/"+name+".sopm", []byte(root.OutputXML(true)), 0644)
			if err != nil {
				panic(err)
			}

			return nil
		},
	}
	unpackCmd.Flags().String("target", "", "target")
	return unpackCmd
}

func buildOpm(file, version string) ([]byte, error) {
	xmlFile, err := os.Open(file)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	// defer the closing of our xmlFile so that we can parse it later on
	defer xmlFile.Close()

	// read our opened xmlFile as a byte array.
	byteValue, err := ioutil.ReadAll(xmlFile)
	if err != nil {
		panic(err)
	}
	doc, err := xmlquery.Parse(bytes.NewBuffer(byteValue))
	if err != nil {
		panic(err)
	}

	err = internal.SOPM2OPM(doc, version)
	if err != nil {
		return nil, err
	}

	return []byte(doc.OutputXML(false)), nil
}

func PostFile(url string, filename string) (*http.Response, error) {
	client := &http.Client{}
	data, err := os.Open(filename)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	return resp, err
}
