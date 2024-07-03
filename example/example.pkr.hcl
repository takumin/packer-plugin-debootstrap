packer {
  required_plugins {
    debootstrap = {
      version = ">= 0.0.1"
      source  = "github.com/takumin/debootstrap"
    }
  }
}

source "debootstrap" "example" {
  suite      = "bookworm"
  variant    = "minbase"
  mirror_url = "http://deb.debian.org/debian"
}

build {
  sources = ["source.debootstrap.example"]

  provisioner "shell" {
    inline = [
      "echo Hello World!",
    ]
  }
}
