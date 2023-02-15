packer {
  required_plugins {
    debootstrap = {
      version = ">= 0.0.1"
      source  = "github.com/takumin/debootstrap"
    }
  }
}

source "debootstrap" "example" {
}

build {
  sources = ["source.debootstrap.example"]
}
