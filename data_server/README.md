# Description
RESTful API for storage and retrieval from the ArrangoDB database

Converting to CSV sample cmd:

    curl --silent -X GET -H "Content-Type: application/json" http://localhost:8529/_db/_system/sensors-data-collector/list?sensor=acc 2>&1 | json2csv -f time,v0,v1,sensor_name,context -o output.csv