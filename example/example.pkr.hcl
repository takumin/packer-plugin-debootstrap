packer {
  required_plugins {
    debootstrap = {
      version = ">= 0.0.1"
      source  = "github.com/takumin/debootstrap"
    }
  }
}

source "debootstrap" "example" {
  suite = "bullseye"
  target_dir = "/tmp/rootfs"
  mirror_url = "http://deb.debian.org/debian"
}

build {
  sources = ["source.debootstrap.example"]
}
