resource "network_request" "network_request" {
  provider      = cidr-reservation
  prefix_length = 26
  base_cidr     = "10.5.0.0/16"
  netmask_id    = "test"
}