package internal

import (
	"encoding/base64"
	"github.com/antchfx/xmlquery"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func SOPM2OPM(doc *xmlquery.Node, version string) error {
	root := xmlquery.FindOne(doc, "//otrs_package")
	if root == nil {
		root = xmlquery.FindOne(doc, "//otobo_package")
	}

	if n := root.SelectElement("//Version"); n != nil {
		if n.FirstChild.Data == "?" {
			n.FirstChild.Data = version
		}
	} else {
		node := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: version,
		}
		pNode := &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "Version",
		}
		xmlquery.AddChild(pNode, node)
		xmlquery.AddChild(root, pNode)
	}

	if n := root.SelectElement("//Vendor"); n != nil {
		if n.FirstChild.Data == "?" {
			n.FirstChild.Data = "FREICON GmbH & Co. KG"
		}
	} else {
		node := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: "FREICON GmbH & Co. KG",
		}
		pNode := &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "Vendor",
		}
		xmlquery.AddChild(pNode, node)
		xmlquery.AddChild(root, pNode)
	}

	if n := root.SelectElement("//URL"); n != nil {
		if n.FirstChild.Data == "?" {
			n.FirstChild.Data = "http://www.freicon.de/"
		}
	} else {
		node := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: "http://www.freicon.de/",
		}
		pNode := &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "URL",
		}
		xmlquery.AddChild(pNode, node)
		xmlquery.AddChild(root, pNode)
	}
	if n := root.SelectElement("//License"); n != nil {
		if n.FirstChild.Data == "?" {
			n.FirstChild.Data = "Copyright by FREICON GmbH & Co. KG"
		}
	} else {
		node := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: "Copyright by FREICON GmbH & Co. KG",
		}
		pNode := &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "License",
		}
		xmlquery.AddChild(pNode, node)
		xmlquery.AddChild(root, pNode)
	}

	if n := root.SelectElement("//BuildDate"); n != nil {
		if n.FirstChild.Data == "?" {
			n.FirstChild.Data = time.Now().Format("2006-01-02 15:04:05")
		}
	} else {
		node := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: time.Now().Format("2006-01-02 15:04:05"),
		}
		pNode := &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "BuildDate",
		}
		xmlquery.AddChild(pNode, node)
		xmlquery.AddChild(root, pNode)
	}

	if n := root.SelectElement("//BuildHost"); n != nil {
		if n.FirstChild.Data == "?" {
			n.FirstChild.Data = "build.freicon.de"
		}
	} else {
		node := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: "build.freicon.de",
		}
		pNode := &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "BuildHost",
		}
		xmlquery.AddChild(pNode, node)
		xmlquery.AddChild(root, pNode)
	}

	var addedFiles []string
	if n := root.SelectElement("//Filelist"); n != nil {
		for _, node := range n.SelectElements("File") {
			b, err := ioutil.ReadFile(node.SelectAttr("Location"))
			if err != nil {
				return errors.WithStack(err)
			}
			addedFiles = append(addedFiles, node.SelectAttr("Location"))

			encoded := base64.StdEncoding.EncodeToString(b)
			xmlquery.AddAttr(node, "Encode", "Base64")
			if node.FirstChild != nil {
				node.FirstChild.Data = encoded
			} else {
				xmlquery.AddChild(node, &xmlquery.Node{
					Type: xmlquery.TextNode,
					Data: encoded,
				})
			}
		}
	} else {
		xmlquery.AddChild(root, &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "Filelist",
		})
	}

	list, err := listFiles(".")
	if err != nil {
		return errors.WithStack(err)
	}

	n := root.SelectElement("//Filelist")
LOOP:
	for _, s := range list {
		for _, addedFile := range addedFiles {
			if addedFile == s {
				continue LOOP
			}
		}

		fileNode := &xmlquery.Node{
			Type: xmlquery.ElementNode,
			Data: "File",
		}
		xmlquery.AddAttr(fileNode, "Location", s)
		xmlquery.AddAttr(fileNode, "Encode", "Base64")
		if strings.HasPrefix(s, "bin") {
			xmlquery.AddAttr(fileNode, "Permission", "755")
		} else {
			xmlquery.AddAttr(fileNode, "Permission", "644")
		}

		b, err := ioutil.ReadFile(s)
		if err != nil {
			return errors.WithStack(err)
		}
		contentNode := &xmlquery.Node{
			Type: xmlquery.TextNode,
			Data: base64.StdEncoding.EncodeToString(b),
		}

		xmlquery.AddChild(fileNode, contentNode)
		xmlquery.AddChild(n, fileNode)
	}
	return nil
}

func listFiles(dir string) ([]string, error) {
	var l []string
	err := filepath.Walk(dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if !strings.HasPrefix(path, "bin/") &&
				!strings.HasPrefix(path, "Custom/") &&
				!strings.HasPrefix(path, "Kernel/") &&
				!strings.HasPrefix(path, "var/") {
				return nil
			}
			l = append(l, path)
			return nil
		})
	return l, err
}
