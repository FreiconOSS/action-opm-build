# OPM build action

This action builds opm files based on .sopm files in your repository. 

## Inputs

## `name`

**Required** The name of the otobo addon you are building.

## `version`

**Required** The version of the he resulting opm.

## `sopm`

The relative path to the .sopm file. If not defined, the action will build `./*.sopm` (i.e. any sopm that is in the root of your repository)


## Example usage

    uses: freiconoss/action-opm-build@v1
    with:
      name: my-otobo-addon
      version: 1.0.0

## Retrieve OPM files
After a successful build, a file called <name>-<version>.opm will be available in the working directory.

It can be collected as an action artifact

    - name: Upload Artifact
      uses: actions/upload-artifact@v3
      with:
        name: my-otobo-addon
        path: my-otobo-addon-1.0.0.opm
        retention-days: 5

Or published as a release artifact

    - name: Get release
      if: always() && github.event_name == 'release'
      id: get_release
      uses: bruceadams/get-release@v1.2.3
      env:
        GITHUB_TOKEN: ${{ github.token }}

    - name: Upload release artifact
      if: always() &&  github.event_name == 'release'
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ github.token }}
      with:
        upload_url: ${{ steps.get_release.outputs.upload_url }}
        asset_path: my-otobo-addon-1.0.0.opm 
        asset_name: my-otobo-addon-1.0.0.opm 
        asset_content_type: application/octet-stream


