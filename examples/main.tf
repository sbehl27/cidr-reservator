resource "cidr-reservator_network_request" "network_request" {
  provider      = cidr-reservator
  prefix_length = 26
  base_cidr     = "10.5.0.0/16"
  netmask_id    = "test"
}