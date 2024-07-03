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
      # update apt repository cache
      "apt-get update",
      # upgrade system
      "apt-get -y dist-upgrade",
    ]
  }

  provisioner "shell" {
    inline = [
      # cleanup apt repository cache
      "find /var/lib/apt/lists -mindepth 1 -maxdepth 1 -type f | xargs rm -f",
      # cleanup deb package cache
      "apt-get clean",
    ]
  }
}
