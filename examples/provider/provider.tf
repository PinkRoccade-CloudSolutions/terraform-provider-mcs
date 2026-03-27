terraform {
  required_providers {
    mcs = {
      source  = "PinkRoccade-CloudSolutions/mcs"
      version = "~> 0.1"
    }
  }
}

variable "mcs_token" {
  type      = string
  sensitive = true
}

provider "mcs" {
  host  = "https://mcs.example.com"
  token = var.mcs_token
}
