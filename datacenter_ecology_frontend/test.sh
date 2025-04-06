#!/usr/bin/env bash

# ----------------------------------------------
# 1) Register a new user
# ----------------------------------------------
echo "=== Registering New User ==="
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "password": "mysecretpassword"
      }'
echo -e "\n"

# Optional: If your server requires a login to set a session cookie, you can do:
# curl -c cookies.txt -X POST http://localhost:8080/login -H "Content-Type: application/json" \
#   -d '{"username":"alice","password":"mysecretpassword"}'
# echo -e "\n"

# ----------------------------------------------
# 2) Add Data Center #1 to Cart
# ----------------------------------------------
echo "=== Adding Data Center #1 to cart ==="
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 45.0,
          "longitude": -93.0,
          "name": "Standard DC",
          "land_price": "$2,000,000",
          "electricity": "$0.07/kWh",
          "notes": "standard"
        },
        "cost": 200000
      }'
echo -e "\n"

# Get Carbon Footprint after adding item #1
echo "=== Getting Carbon Footprint (after #1) ==="
curl "http://localhost:8080/cart/carbon-footprint?username=alice"
echo -e "\n"

# ----------------------------------------------
# 3) Add Data Center #2 to Cart
# ----------------------------------------------
echo "=== Adding Data Center #2 to cart ==="
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 40.7,
          "longitude": -74.0,
          "name": "Eco Optimized Center",
          "land_price": "$3,500,000",
          "electricity": "Renewable",
          "notes": "eco"
        },
        "cost": 350000
      }'
echo -e "\n"

# Get Carbon Footprint after adding item #2
echo "=== Getting Carbon Footprint (after #2) ==="
curl "http://localhost:8080/cart/carbon-footprint?username=alice"
echo -e "\n"

# ----------------------------------------------
# 4) Add Data Center #3 to Cart
# ----------------------------------------------
echo "=== Adding Data Center #3 to cart ==="
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 37.77,
          "longitude": -122.42,
          "name": "Next-Gen Sustainable Facility",
          "land_price": "$5,000,000",
          "electricity": "$0.10/kWh",
          "notes": "next-gen"
        },
        "cost": 500000
      }'
echo -e "\n"

# Get Carbon Footprint after adding item #3
echo "=== Getting Carbon Footprint (after #3) ==="
curl "http://localhost:8080/cart/carbon-footprint?username=alice"
echo -e "\n"

