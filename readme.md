# Building Location and Trip Planner Service (Part ||)
## Building Trip planner 
The trip planner is a feature that will take a set of locations from the database and will then check against UBER’s price estimates API to suggest the best possible route in terms of costs and duration.


To access Uber Sandbox API, OAuth token is generated and captured by redirecting user to http://localhost:7635. For authenticating user the first time, user is redirected to Uber site in the browser.
This OAuth access token is for single use and expires in 10 minutes.

```
export GOPATH=$PWD # Set GO PATH
go get github.com/fighterlyt/permutation #Get dependencies
go get github.com/julienschmidt/httprouter #Get dependencies
go get gopkg.in/tomb.v2 #Get dependencies
go get gopkg.in/mgo.v2 #Get dependencies
go build -v src/server/Server.go  # Build server
```

## Run the server using 

```
go run src/server/Server.go
```

or

```
./Server
```

## API Calls

### Plan a trip

For planning send request to server using following:

```
curl -v -X POST -H 'Content-Type:application/json' -d '{"starting_from_location_id": "561b2eba5cb8b322702fbe82", "location_ids" : [ "561b304d5cb8b322702fbe84", "561b30835cb8b322702fbe85", "561b2f365cb8b322702fbe83", "561b31435cb8b322702fbe86" ] }' http://localhost:7000/trips
```

Sample response is : HTTP Response CODE : 201

```
{"id":"564eb94e1c0296601edd87e3","status":"planning","starting_from_location_id":"561b2eba5cb8b322702fbe82","best_route_location_ids":["561b31435cb8b322702fbe86","561b2f365cb8b322702fbe83","561b30835cb8b322702fbe85","561b304d5cb8b322702fbe84"],"total_uber_costs":56,"total_uber_duration":4289,"total_distance":30.290000000000003}
```

### Check trip details and status 

Following request to server returns trip details and status

```
curl -X GET -v http://localhost:7000/trips/564eb94e1c0296601edd87e3
```
where 564eb94e1c0296601edd87e3 is trip_id obtained in previous request as id

Sample response is : HTTP Response CODE : 200

```
{"id":"564eb94e1c0296601edd87e3","status":"planning","starting_from_location_id":"561b2eba5cb8b322702fbe82","best_route_location_ids":["561b31435cb8b322702fbe86","561b2f365cb8b322702fbe83","561b30835cb8b322702fbe85","561b304d5cb8b322702fbe84"],"total_uber_costs":56,"total_uber_duration":4289,"total_distance":30.290000000000003} 
```

### Start(Or Update) a Trip

To start a trip send the following request to server
```
curl -X PUT -v http://localhost:7000/trips/564eb94e1c0296601edd87e3/request
``` 

Sample response is : HTTP Response CODE : 201
```
{"id":"564eb94e1c0296601edd87e3","status":"requesting","starting_from_location_id":"561b2eba5cb8b322702fbe82","next_destination_location_id":"561b31435cb8b322702fbe86","best_route_location_ids":["561b31435cb8b322702fbe86","561b2f365cb8b322702fbe83","561b30835cb8b322702fbe85","561b304d5cb8b322702fbe84"],"total_uber_costs":56,"total_uber_duration":4289,"total_distance":30.290000000000003,"uber_wait_time_eta":7} 
```

The above calls start trip from location 561b2eba5cb8b322702fbe82 to 561b31435cb8b322702fbe86

Subsequent calls to same endpoint will pull up destination from the best route location ids list in order.

```
curl -XPUT -v http://localhost:7000/trips/564eb94e1c0296601edd87e3/request
```

Sample response is : HTTP Response CODE : 201

```
{"id":"564eb94e1c0296601edd87e3","status":"requesting","starting_from_location_id":"561b2eba5cb8b322702fbe82","next_destination_location_id":"561b2f365cb8b322702fbe83","best_route_location_ids":["561b31435cb8b322702fbe86","561b2f365cb8b322702fbe83","561b30835cb8b322702fbe85","561b304d5cb8b322702fbe84"],"total_uber_costs":56,"total_uber_duration":4289,"total_distance":30.290000000000003,"uber_wait_time_eta":7}
```

Once all destination ids are processed (along with making the last PUT call made to 'starting location' to make it as round trip), 
the status of trip is changed to 'finished'. 
Any further PUT requests for this trip would throw 404 error as whole Trip has already been processed.

