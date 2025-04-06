#!/usr/bin/env bash

# Adds 5 different data centers to Alice's cart
# Then gets the simulation output

echo "Adding Data Center #1 to cart..."
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 39.0438,
          "longitude": -77.4874,
          "name": "AWS US East Region (Ashburn VA)",
          "land_price": "1.5-2.5M per acre",
          "electricity": "$0.06-0.08/kWh",
          "notes": "AWS largest and oldest region in Data Center Alley"
        },
        "cost": 250000
      }'
echo -e "\n"

echo "Adding Data Center #2 to cart..."
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 38.9661,
          "longitude": -77.3682,
          "name": "Equinix DC1 (Ashburn VA)",
          "land_price": "1.5-2.5M per acre",
          "electricity": "$0.07-0.09/kWh",
          "notes": "Major internet exchange point in Northern Virginia"
        },
        "cost": 300000
      }'
echo -e "\n"

echo "Adding Data Center #3 to cart..."
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 38.9654,
          "longitude": -77.3591,
          "name": "Microsoft Azure East US (Boydton VA)",
          "land_price": "1-2M per acre",
          "electricity": "$0.06-0.08/kWh",
          "notes": "One of Microsofts largest data center regions"
        },
        "cost": 150000
      }'
echo -e "\n"

echo "Adding Data Center #4 to cart..."
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 38.7841,
          "longitude": -77.1710,
          "name": "Iron Mountain VA-1 (Manassas VA)",
          "land_price": "800K-1.2M per acre",
          "electricity": "$0.06-0.08/kWh",
          "notes": "LEED Gold certified facility"
        },
        "cost": 200000
      }'
echo -e "\n"

echo "Adding Data Center #5 to cart..."
curl -X POST http://localhost:8080/cart/add \
  -H "Content-Type: application/json" \
  -d '{
        "username": "alice",
        "item": {
          "latitude": 37.2665,
          "longitude": -79.9413,
          "name": "QTS Richmond (Richmond VA)",
          "land_price": "400-600K per acre",
          "electricity": "$0.06-0.08/kWh",
          "notes": "Former semiconductor plant converted to data center"
        },
        "cost": 100000
      }'
echo -e "\n"

echo "Retrieving simulation results for alice..."
curl "http://localhost:8080/api/simulation?username=alice"
echo -e "\n"

