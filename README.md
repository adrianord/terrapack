# Terrapack

**Terraform Pack** is a small tool for packing and uploading Terraform Configurations to Terraform Enterprise.
Terrapack was built out of necessity to run [cdktf](https://github.com/hashicorp/terraform-cdk) stacks on Terraform
Enterprise.

## Installing

```bash
$ curl https://github.com/adrianord/terrapack/releases/download/v0.1.0/terrapack_0.1.0_Linux_x86_64.tar.gz -Lo terrapack.tar.gz
$ tar -zxvf terrapack.tar.gz
$ sudo mv terrapack /usr/local/bin
```

## Example usages

Terrapack will try to auto discover the remote backend from the `terraform` block.

```terraform
terraform {
  backend "remote" {
    hostname = "app.terraform.io"
    organization = "myorganization"

    workspaces {
      name = "myworkspace"
    }
  }
}
```

```bash
$ terrapack terraform_directory/
```

> **Note:** Terrapack does not contact Terraform Enterprise to resolve the remote backend information. Because of this,
Terrapack does not support `prefix` and only takes the first `workspaces` block.

If Terrapack cannot find the information needed, flags are available.

```bash
$ terrapack -u app.terraform.io -o myorganization example/
```

Terrapack uses the [go-tfe](https://github.com/hashicorp/go-tfe) Go module to support
the usual environment variables (TFE_TOKEN, etc...), as well as the `.terraformignore`
file.

## cdktf

This tool was made to work with cdktf. It allows you to use cdktf with a gitops approach without leaving artifacts in the repository.

```bash
cdktf synth
terrapack cdktf.out/stacks/examplestack/
```
