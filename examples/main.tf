resource "cidr-reservator_network_request" "network_request" {
  provider      = cidr-reservator
  prefix_length = 26
  base_cidr     = "10.5.0.0/16"
  netmask_id    = "test"
}

resource "cidr-reservator_network_request" "network_request2" {
  provider      = cidr-reservator
  prefix_length = 26
  base_cidr     = "10.6.0.0/18"
  netmask_id    = "test"
}