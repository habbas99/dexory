# Dexory

## The task
Our robot systems scan a customer warehouse automatically every evening and produce a
report describing the location of assets. This information is made available as a JSON file, which
a robot automatically uploads to an HTTP endpoint at the end of a scan. This JSON file consists
of a list of all locations that were scanned in the warehouse, and indicates if they were occupied
during the scan, along with any barcodes that were scanned in the location.
Separately, the customer will upload a CSV file containing a list of the items that they expect to
have been located in each location in the warehouse during the scan.
The requirement is to provide an application which can compare the data gathered by the robot
and the data supplied by the customer, and generate a report that identifies any discrepancies.

The output should provide a result for comparing each location, and consist of:

- The name of the location
- Whether or not the location was successfully scanned
- Whether or not the location was occupied
- The barcodes that were expected to be found in this location
- The barcodes that were actually found in this location
- A description of the outcome of the comparison, using at least the following statuses:
    - The location was empty, as expected
    - The location was empty, but it should have been occupied
    - The location was occupied by the expected items
    - The location was occupied by the wrong items
    - The location was occupied by an item, but should have been empty
    - The location was occupied, but no barcode could be identified

## Functional requirements
- A web application is available which can consume JSON from a robot and store it in an appropriate data store.
- A user can select a previously-uploaded JSON payload, and upload a CSV report to generate a comparison report.
- A user can view or export the comparison report in an appropriate format.
- The JSON and CSV upload endpoints accept the format supplied in the sample files accompanying this coding exercise.

## Non-functional requirements
- You should include appropriate documentation about your code.
- Your code should follow best practices and a consistent coding style.
- Your submission should include whatever test cases you consider appropriate.


## Setup

### Pre-requisites
*Note*: Make sure that you are in the root folder.

GoLang version: `go1.23.0 linux/amd64`

NodeJS version: `v18.15.0`

Start postgres in docker: 
```
docker-compose up -d
```

Stop postgres in docker:
```
docker-compose down
```

### Testing
Note: mocks are commited in `generated` folder

Sample command to generate mocks:
```
mockgen -source=internal/services/export/export_report_service.go -destination=generated/services/export/mock_export_report_service_interfaces.go -package=mockexportreportservice
```

Run all tests:
```
go test -v ./...
```

### Backend server
```
go run cmd/dexory/main.go
```

### Frontend application
```
npm install

npm start
```

### Application usage
Make sure to change `REPLACE_ME` in `curl` command with path to sample JSON file with scans.

API to upload scans from robot:
```
curl -X POST http://localhost:8080/upload-bulk-scan-file -F "file=@{REPLACE_ME}/example-customer.json"
```

Access development frontend application: http://localhost:3000

To generate comparison report, navigate to frontend. Once report is generated there is an option to export the report in JSON format.

Sample exported report can be found under this path: `/sample/report.json`

### Production build and usage
Update environment variable `ENVIRONMENT` to `production` in `.env` file

Build application:
```
go build -o "./build/dexory" -v "./cmd/dexory"
```

Run application:
```
build/dexory
```

Note: frontend is packaged with the backend server

Access frontend: http://localhost:8080

## Assumptions
- multiple barcodes could be received from robot for a given location
- report contains expected and actual barcodes as an array even though CSV only has a single barcode in each row
- allow robot to send duplicate scans `JSON` file
- allow user to upload same `CSV` file to generate report against same robot scans file

## Future considerations
- protect upload bulk scans file API from duplicate file content by storing hash
- generate report API from generating comparison data that has already been generated (store hash of CSV file)
- add more test coverage including unit and integration tests for frontend/backend
- support pagination and filtering on report detail view
- support report summary generation on backend as opposed to frontend
- use custom errors application generated errors