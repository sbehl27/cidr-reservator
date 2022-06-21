# Improvements

- read all baseCidr files and check by their name, that the base ranges do not overlap!
- check that a network request prefix can never be bigger (or even as big) as the full base cidr range!
- allow array of baseCidrs! if first is exhausted, use second!
- upload the list of reserved cidr ranges sorted!