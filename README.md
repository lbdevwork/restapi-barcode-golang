# RestAPI Barcode GoLang

A RESTful API to fetch product information using barcodes, written in Go. The API returns information about products in JSON or plain text format, with support for English (en) and Portuguese (pt) languages.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [License](#license)

## Installation

1. Clone this repository:

        git clone https://github.com/lbdevwork/restapi-barcode-golang.git 

2. Change into the project directory:

        cd restapi-barcode-golang

3. Install the required dependencies:

        go get -u ./...

4. Build and run the application:

        go build -o app cmd/barcode_scanner/main.go
        ./app

The API will be available at http://localhost:8080/v1.


## Usage
The API accepts requests with product barcodes and returns information in JSON or plain text format. The responses include details about the product, such as name, Nutriscore grade, Ecoscore grade, and nutritional information.

## API Endpoints
GET /product/{barcode}
Fetch product information in JSON format

GET /product/text/lang/{barcode}
Fetch product information in plain text format, with support for "en" (English) and "pt" (Portuguese) languages. Replace {barcode} with the product's barcode.

## Production
Google Cloud Run Deployment on branch : google-run-microservice

## License
This project is licensed under the MIT License - see the LICENSE file for details.


