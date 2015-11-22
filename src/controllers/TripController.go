package controllers

import (  
    "encoding/json"
    "fmt"
    "net/http"
	"gopkg.in/mgo.v2"
    "github.com/julienschmidt/httprouter"
    "gopkg.in/mgo.v2/bson"
    "model"
    "io/ioutil"
    "strconv"
    "github.com/fighterlyt/permutation"
)

/*
* TripController represents the controller for operating on the Trip resource
*/
type (  
    TripController struct {  
    	session *mgo.Session
	}
)

func NewTripController(s *mgo.Session) *TripController {  
    return &TripController{s}
}

/*
* GetAllTrips - retrieves all Trip resources stored in mongolab
*/
func (lc TripController) GetAllTrips(w http.ResponseWriter, r *http.Request, p httprouter.Params) {  
    // Stub Trips
    var Trips []model.TripResponse

    // Fetch Trips
    if err := lc.session.DB("cmpe273-assignment2").C("trips").Find(bson.M{}).All(&Trips); err != nil {
        w.WriteHeader(404)
        return
    }

	results := ""
	for index := 0; index < len(Trips); index++ {
	    // Marshal provided interface into JSON structure
	    TripJson, _ := json.Marshal(Trips[index])
		results = results + string(TripJson) + "\n "
	}
	// Write content-type, statuscode, payload
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	fmt.Fprintf(w, "%s", results)
	fmt.Println("All Trip resources retrieved successfully from MongoLab!")
	fmt.Println("------------------------------------------------------------")	
}


/*
* GetTrip - retrieves an individual Trip resource (based on id) stored in mongolab
*/
func (lc TripController) GetTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {  
    // Grab id
    id := p.ByName("id")

    // Verify id is ObjectId, otherwise bail
    if !bson.IsObjectIdHex(id) {
        w.WriteHeader(404)
        return
    }

    // Grab id
    oid := bson.ObjectIdHex(id)
    // Stub Trip
    Trip := model.TripResponse{}

    // Fetch Trip
    if err := lc.session.DB("cmpe273-assignment2").C("trips").FindId(oid).One(&Trip); err != nil {
        w.WriteHeader(404)
        return
    }

    // Marshal provided interface into JSON structure
    TripJson, _ := json.Marshal(Trip)

    // Write content-type, statuscode, payload
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(200)
    fmt.Fprintf(w, "%s", TripJson)
	fmt.Println("Trip resource retrieved successfully from MongoLab!")
	fmt.Println("------------------------------------------------------------")
}

/*
* CreateTrip - creates an individual Trip resource and stores it in mongolab
*/
func (lc TripController) CreateTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {  
    // Stub a Trip to be populated from the body
    request := model.TripRequest{}
    response := model.TripResponse{}

    // Populate the Trip data
    json.NewDecoder(r.Body).Decode(&request)

    // Add an Id
    response.Id = bson.NewObjectId()
    response.StartingFromLocationId = request.StartingFromLocationId

	//Call Uber Price Estimates API
	fieldsMissing := FindMissingFields(request)
	if(fieldsMissing == true) {
		    newErrorResponse := model.Error{}
		    newErrorResponse.Code = "MISSING_REQUIRED_FIELDS"
		    newErrorResponse.Description = "Location fields are empty. Please enter valid locations."
		    newErrorResponse.Fieldname="StartingFrom Location ID,Destination Location IDs"
		    newErrorJson, _ := json.Marshal(newErrorResponse)
		    	
		    w.Header().Set("Content-Type", "application/json")
		    w.WriteHeader(400)
		    fmt.Fprintf(w, "%s", newErrorJson)	
	} else {
		//GET the latitude and longitude for Starting from location IDs and Destination IDs
		var startLatitudeOrig, startLongitudeOrig float64
		var invalidLocationDetails bool
		var errorCode, errorDescription string
		storedStartingLocationID := request.StartingFromLocationId
		startLatitudeOrig, startLongitudeOrig, invalidLocationDetails, errorCode, errorDescription = GetLocationCoordinatesFromMongoDB(lc, request.StartingFromLocationId)
		if(invalidLocationDetails == true) {
		    newErrorResponse := model.Error{}
		    newErrorResponse.Code = errorCode
		    newErrorResponse.Description = errorDescription
		    newErrorResponse.Fieldname = errorDescription
		    newErrorJson, _ := json.Marshal(newErrorResponse)
		    	
		    w.Header().Set("Content-Type", "application/json")
		    w.WriteHeader(400)
		    fmt.Fprintf(w, "%s", newErrorJson)	
		} else {
			//Get Location Coordinates for Destination IDs too
			var endLatitudes = make([]float64, len(request.LocationIds))
			var endLongitudes = make([]float64, len(request.LocationIds))
			
			//Find out the possible permutations (routes)
			var possibleRoutes = Factorial(len(request.LocationIds))
			fmt.Println("No of Possible Routes with the given destination IDs : ", possibleRoutes)
			possiblePermutations := Permutations(request.LocationIds, possibleRoutes)
			
			var totalUberCost = make([]float64, possibleRoutes)
			var totalUberDuration = make([]float64, possibleRoutes)
			var totalDistance = make([]float64, possibleRoutes)
			//var CalculatedCosts = make([][5]string, possibleRoutes*len(request.LocationIds))
			CalculatedCosts := make(map[string][]float64)
			
			//Variable declarations for storing the least values for Uber cost among all the possible routes
			var leastUberCostAmongAll, leastUberDuration,leastUberDistance float64
			leastUberCostAmongAll = 999999.99 
			leastUberDuration = 999999.99
			leastUberDistance = 999999.99
			var bestRouteLocationIDs = make([]string, len(request.LocationIds))
			//calcIndex := 0
			var uberCost, uberDuration, uberDistance float64
			var invalidAddress bool
			var errorCode, errorMsg string
			
			for i := 0; i < len(possiblePermutations); i++ {
				startLatitude := startLatitudeOrig
				startLongitude := startLongitudeOrig
				
				for index := 0; index < len(possiblePermutations[i]); index++ {
					endLatitudes[index], endLongitudes[index], _, _, _ = GetLocationCoordinatesFromMongoDB(lc, possiblePermutations[i][index])
					
					//Check if this cost was calculated and saved earlier
					isAlreadyCalculated, whichIdx := IsAlreadyCalculated(CalculatedCosts, storedStartingLocationID,possiblePermutations[i][index])
					if(isAlreadyCalculated == true) {
						uberCost = CalculatedCosts[whichIdx][0]
						uberDuration = CalculatedCosts[whichIdx][1]
						uberDistance = CalculatedCosts[whichIdx][2]
					} else {
						uberCost, uberDuration, uberDistance, invalidAddress, errorCode, errorMsg = CallUberAPIForPriceEstimates(startLatitude, startLongitude, endLatitudes[index], endLongitudes[index])
						mapId := storedStartingLocationID + "-" + possiblePermutations[i][index]
						CalculatedCosts[mapId] = append(CalculatedCosts[mapId], uberCost)
						CalculatedCosts[mapId] = append(CalculatedCosts[mapId], uberDuration)
						CalculatedCosts[mapId] = append(CalculatedCosts[mapId], uberDistance)
					}

					//Change start coordinates to the next location
					startLatitude = endLatitudes[index]
					startLongitude = endLongitudes[index]
					storedStartingLocationID = possiblePermutations[i][index]
					
					if(invalidAddress) {	
					    newErrorResponse := model.Error{}
					    newErrorResponse.Code = errorCode
					    newErrorResponse.Description = errorMsg
					    newErrorResponse.Fieldname=""
			
					    newErrorJson, _ := json.Marshal(newErrorResponse)
					    	
					    w.Header().Set("Content-Type", "application/json")
					    w.WriteHeader(400)
					    fmt.Fprintf(w, "%s", newErrorJson)
						
					} else {			
						//add these values
						//get totals
						totalUberCost[i] = totalUberCost[i] + uberCost
						totalUberDuration[i] = totalUberDuration[i] +  uberDuration
						totalDistance[i] = totalDistance[i] + uberDistance
				    }					
				}
				
				//Add one pricing estimate for round trip too					
				uberCost, uberDuration, uberDistance, invalidAddress, errorCode, errorMsg := CallUberAPIForPriceEstimates(startLatitude, startLongitude, startLatitudeOrig, startLongitudeOrig)

				if(invalidAddress) {	
					    newErrorResponse := model.Error{}
					    newErrorResponse.Code = errorCode
					    newErrorResponse.Description = errorMsg
					    newErrorResponse.Fieldname=""
			
					    newErrorJson, _ := json.Marshal(newErrorResponse)
					    	
					    w.Header().Set("Content-Type", "application/json")
					    w.WriteHeader(400)
					    fmt.Fprintf(w, "%s", newErrorJson)
						
				} else {			
						//add these values
						//get totals
						totalUberCost[i] = totalUberCost[i] + uberCost
						totalUberDuration[i] = totalUberDuration[i] +  uberDuration
						totalDistance[i] = totalDistance[i] + uberDistance
				}					
					
				if(totalUberCost[i] < leastUberCostAmongAll) {
							leastUberCostAmongAll = totalUberCost[i]
							leastUberDuration = totalUberDuration[i]
							leastUberDistance = totalDistance[i]
							bestRouteLocationIDs = possiblePermutations[i]
				} else if(totalUberCost[i] == leastUberCostAmongAll && totalUberDuration[i] < leastUberDuration) {
							leastUberCostAmongAll = totalUberCost[i]
							leastUberDuration = totalUberDuration[i]
							leastUberDistance = totalDistance[i]
							bestRouteLocationIDs = possiblePermutations[i]
				}				
			} 
			fmt.Println("The Best Route Location IDs for this trip would be : ",bestRouteLocationIDs)
			fmt.Println("The least Uber Cost for these destinations would be : ",leastUberCostAmongAll)
			fmt.Println("The least Uber Duration for these destinations would be : ",leastUberDuration)
			fmt.Println("The least Uber Distance for these destinations would be : ",leastUberDistance)
			fmt.Println("=================================================")
			
			//Write the Trip to mongo
			response.Status = "planning"
			response.BestRouteLocationIds = bestRouteLocationIDs
			response.TotalUberCosts = leastUberCostAmongAll
			response.TotalUberDuration = leastUberDuration
			response.TotalDistance = leastUberDistance
			lc.session.DB("cmpe273-assignment2").C("trips").Insert(response)
			fmt.Println("Trip data written successfully to MongoLab!")
			fmt.Println("------------------------------------------------------------")		
					
			// Marshal provided interface into JSON structure
			responseJson, _ := json.Marshal(response)
			
			// Write content-type, statuscode, payload
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(201)
			fmt.Fprintf(w, "%s", responseJson)
			
		}
    }
}

/*
* UpdateTrip - updates an individual Trip resource (based on id) stored in mongolab
*/
//Variables to keep track of current destination ID being reached
var currentDestinationID string
var currentStartLocationID string
var currentIndex int
var originalStartingLocationID string
//To keep track of the pointer for any processed trips
var processedTrips = make(map[string]int) 

func (lc TripController) UpdateTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//Request for new oid
	if(currentIndex == -1) {
		currentIndex = 0
	}
	
    // Grab id
    id := p.ByName("id")
	
	indexForThisTrip := processedTrips[id]
	if(indexForThisTrip == 0) {
		currentIndex = 0
		fmt.Println("Starting afresh this trip!")
	} else {
		currentIndex = indexForThisTrip
		fmt.Println("Looks like this trip was processed earlier. We would continue looking for the unprocessed destination points!")
	}
	
	
    // Verify id is ObjectId, otherwise bail
    if !bson.IsObjectIdHex(id) {
        w.WriteHeader(404)
        return
    }

    // Grab id
    oid := bson.ObjectIdHex(id)

    // Stub Trip
    Trip := model.TripUpdateResponse{}

	collection := lc.session.DB("cmpe273-assignment2").C("trips")
    // Fetch Trip
    if err := collection.FindId(oid).One(&Trip); err != nil {
        w.WriteHeader(404)
        return
    }

	if(Trip.Status != "finished") {
		var status string
		
		if(currentIndex == 0) { //Start from Origin to first destination location ID {
			originalStartingLocationID = Trip.StartingFromLocationId
			currentStartLocationID = Trip.StartingFromLocationId
			currentDestinationID = Trip.BestRouteLocationIds[currentIndex]
			fmt.Println("Traveling from : ",currentStartLocationID," To ",currentDestinationID)		
		} else if(currentIndex == len(Trip.BestRouteLocationIds)) {
			fmt.Println("Let's go back to home!!")
			currentStartLocationID = Trip.BestRouteLocationIds[len(Trip.BestRouteLocationIds) - 1]
			currentDestinationID = originalStartingLocationID
			status = "finished"
			fmt.Println("Traveling from : ",currentStartLocationID," To ",currentDestinationID)
		} else if(currentIndex == -1) {
			fmt.Println("This trip is finished!!")
			
		} else {
			if(currentIndex < len(Trip.BestRouteLocationIds)) {
				currentStartLocationID = Trip.BestRouteLocationIds[currentIndex - 1]
				currentDestinationID = Trip.BestRouteLocationIds[currentIndex]
				fmt.Println("Traveling from : ",currentStartLocationID," To ",currentDestinationID)
			}
		}
		
		var rideResponse model.RideRequestResponse
		
		
		if(currentIndex != -1) {
					status = "requesting"
					rideResponse = CallUberAPIForRideRequest(lc, currentStartLocationID, currentDestinationID) 
						
				    if(currentIndex == len(Trip.BestRouteLocationIds)) {
				    	processedTrips[id] = currentIndex;
				    	currentIndex = -1
				    	status = "finished"
				    } else {
				    	currentIndex++
				    	processedTrips[id] = currentIndex;
				    }
		} 
		
		// Fetch location
		oldTrip := bson.M{"id": oid}
		if err := collection.FindId(oid).One(&oldTrip); err != nil {
		       w.WriteHeader(404)
		       return
		}
		
		newTrip := bson.M{"$set": bson.M{"status": status , "starting_from_location_id": Trip.StartingFromLocationId, "next_destination_location_id": currentDestinationID, 
										"BestRouteLocationIds": Trip.BestRouteLocationIds, 
										"total_uber_costs": Trip.TotalUberCosts, "total_uber_duration" : Trip.TotalUberDuration , "total_distance" : Trip.TotalDistance,
										"uber_wait_time_eta" : rideResponse.Eta }}
		err := collection.Update(oldTrip, newTrip)
		if err != nil {
				fmt.Println("Error : Error received while updating trip : ",err)
				fmt.Println("------------------------------------------------------------")					
		} else {   
					    //Fetch new location
					    
					    newTripResponse := model.TripUpdateResponse{}	
					    if err := collection.FindId(oid).One(&newTripResponse); err != nil {
					        fmt.Println("Error : Error received while fetching updated trip!: ",err)
					    }
					
					    // Marshal provided interface into JSON structure
					    newTripJson, _ := json.Marshal(newTripResponse)
					
					    // Write content-type, statuscode, payload
					    w.Header().Set("Content-Type", "application/json")
					    w.WriteHeader(201)
					    fmt.Fprintf(w, "%s", newTripJson)
						fmt.Println("Trip resource updated successfully in MongoLab!")
						fmt.Println("------------------------------------------------------------")				    
		}

	} else {
		fmt.Println("This trip has already been processed. The status is 'finished'!")
		w.WriteHeader(404)
        return
	}
		
}

/*
* RemoveTrip - removes an individual Trip resource (based on id) stored in mongolab
*/
func (lc TripController) RemoveTrip(w http.ResponseWriter, r *http.Request, p httprouter.Params) {  
    // Grab id
    id := p.ByName("id")

    // Verify id is ObjectId, otherwise bail
    if !bson.IsObjectIdHex(id) {
        w.WriteHeader(404)
        return
    }

    // Grab id
    oid := bson.ObjectIdHex(id)

    // Remove user
    if err := lc.session.DB("cmpe273-assignment2").C("trips").RemoveId(oid); err != nil {
        w.WriteHeader(404)
        return
    }

    // Write status
    w.WriteHeader(200)
    fmt.Println("Trip resource removed successfully from MongoLab!")
	fmt.Println("------------------------------------------------------------")    
}


/**
* Utility Functions
*/
func FindMissingFields(request model.TripRequest) bool {
	//Grab and store request params in strings
	StartingFromLocationId := request.StartingFromLocationId
	LocationIds := request.LocationIds
	if(StartingFromLocationId == "" && len(LocationIds)==0) {
		fmt.Println("Error : Request Parameters are empty!")
		fmt.Println("------------------------------------------------------------")
		return true
	}
	
	return false
}

//Function to check if any Price Estimate call for any startLocation-to-endLocation is already available in the Map
func IsAlreadyCalculated(CalculatedCosts map[string][]float64, StartingLocationID string, DestinationLocationID string) (bool, string) {
	mapid := StartingLocationID + "-" + DestinationLocationID
	dataInMapForThisKey := CalculatedCosts[mapid]
	if(len(dataInMapForThisKey) == 0) {
		return false, ""
	} else {
		return true, mapid
	}
	
}

//Function to find out no. of possible permuations with given destination IDs
func Factorial(noOfIDs int) int {
	var noOfRoutes int
	noOfRoutes = 1;
	for index := noOfIDs; index > 0; index-- {
		noOfRoutes = noOfRoutes * index	
	}
	return noOfRoutes
}


//Find out all possible permutation with given Destination Location IDs
func Permutations(LocationIds []string, possibleRoutes int) [][]string {        
        var possiblePermutations = make([][]string, possibleRoutes)
		i := LocationIds
        p,err := permutation.NewPerm(i,nil) //generate a Permutator
        if err != nil {
            fmt.Println(err)
            return nil
        }
        for i,err:=p.Next();err==nil;i,err=p.Next(){
            //fmt.Printf("%3d permutation: %v left %d\n",p.Index()-1,i.([]string),p.Left())
            possiblePermutations[p.Index()-1] = i.([]string)
        }        
        
        return possiblePermutations
}

//Function to get Location Coordinates Saved in Mongodb
func GetLocationCoordinatesFromMongoDB(lc TripController, StartingFromLocationId string) (float64, float64, bool, string, string) {
	var invalidLocationCoordinateDetails bool
	invalidLocationCoordinateDetails = false
	var errorDescription string
	var errorCode string

	// Stub location
    location := model.Location{}

    // Verify id is ObjectId, otherwise bail
    if !bson.IsObjectIdHex(StartingFromLocationId) {
        invalidLocationCoordinateDetails = true
		errorCode = "GET_LOCATION_COORDINATES_API_ERROR"
		errorDescription = "Error received while fetching Location Coordinate from Mongodb. Please check the location details."
		fmt.Println("Error : Error received while fetching Location Coordinate from Mongodb! : ");
		fmt.Println("------------------------------------------------------------")
		return 0.0, 0.0, invalidLocationCoordinateDetails, errorCode, errorDescription
    }

    // Grab id
    oid := bson.ObjectIdHex(StartingFromLocationId)

    // Fetch location //"cmpe273-assignment2"
    if err := lc.session.DB("cmpe273-assignment2").C("locations").FindId(oid).One(&location); err != nil {
        invalidLocationCoordinateDetails = true
		errorCode = "GET_LOCATION_COORDINATES_API_ERROR"
		errorDescription = "Error received while fetching Location Coordinate from Mongodb. Please check the location details."
		fmt.Println("Error : Error received while fetching Location Coordinate from Mongodb! : ", err);
		fmt.Println("------------------------------------------------------------")
		return 0.0, 0.0, invalidLocationCoordinateDetails, errorCode, errorDescription
    }
	
	var latitude, longitude float64
	latitude = location.Coordinate.Latitude
	longitude = location.Coordinate.Longitude		

	startLatitude := strconv.FormatFloat(latitude, 'f', -1, 32)
	startLongitude := strconv.FormatFloat(longitude, 'f', -1, 32)
	
	if(startLongitude != "" && startLatitude != ""){
		errorDescription = ""
	} else {
		errorCode = "INVALID_LOCATIOn"
		errorDescription = "No results returned! Please enter valid Starting location ID and Destination IDs"
		fmt.Println("Error : No results returned! Please enter valid Starting location ID and Destination IDs")
		fmt.Println("------------------------------------------------------------")
		invalidLocationCoordinateDetails = true
		return 0, 0, invalidLocationCoordinateDetails, errorCode, errorDescription
	}
	
	return latitude, longitude, invalidLocationCoordinateDetails, errorCode, errorDescription
}

//Function to get Location Coordinates By Making 'GET' call to Location Service(Assignment 2)
func GetLocationCoordinates(StartingFromLocationId string) (float64, float64, bool, string, string) {
	var query string
	var invalidLocationCoordinateDetails bool
	invalidLocationCoordinateDetails = false
	var errorDescription string
	var errorCode string
	
	errorDescription = ""
	errorCode = ""
	query = "http://localhost:5000/locations/" + StartingFromLocationId
	resp, err := http.Get(query)
	
	if err != nil {
		invalidLocationCoordinateDetails = true
		errorCode = "GET_LOCATION_COORDINATES_API_ERROR"
		errorDescription = "Error received while executing Location Coordinate API call. Please check the location details."
		fmt.Println("Error : Error received while executing Location Coordinate API call! : ", err);
		fmt.Println("------------------------------------------------------------")
		return 0.0, 0.0, invalidLocationCoordinateDetails, errorCode, errorDescription
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		errorCode = "GET_LOCATION_COORDINATES_API_ERROR"
		errorDescription = "Unable to read response from Location Coordinate API. Please try again later."
		fmt.Println("Error : Error received while reading response from Location Coordinate API call! : ", err);
		fmt.Println("------------------------------------------------------------")	
		invalidLocationCoordinateDetails = true
		return 0, 0, invalidLocationCoordinateDetails, errorCode, errorDescription
	}
		
	var response model.Location
	errUnmarshal := json.Unmarshal(body, &response)
	if errUnmarshal != nil {
		errorCode = "GET_LOCATION_COORDINATES_API_ERROR"
		errorDescription = "Unable to unmarshall response from Location Coordinate API. Please try again later."
		fmt.Println("Error : Error received while unmashalling response from Location Coordinate API call!: ", errUnmarshal);
		fmt.Println("------------------------------------------------------------")	
		invalidLocationCoordinateDetails = true
		return 0, 0, invalidLocationCoordinateDetails, errorCode, errorDescription
	}
		
	var latitude, longitude float64
	latitude = response.Coordinate.Latitude
	longitude = response.Coordinate.Longitude		

	startLatitude := strconv.FormatFloat(latitude, 'f', -1, 32)
	startLongitude := strconv.FormatFloat(longitude, 'f', -1, 32)
	
	if(startLongitude != "" && startLatitude != ""){
		errorDescription = ""
	} else {
		errorCode = "INVALID_LOCATIOn"
		errorDescription = "No results returned! Please enter valid Starting location ID and Destination IDs"
		fmt.Println("Error : No results returned! Please enter valid Starting location ID and Destination IDs")
		fmt.Println("------------------------------------------------------------")
		invalidLocationCoordinateDetails = true
		return 0, 0, invalidLocationCoordinateDetails, errorCode, errorDescription
	}
	
	return latitude, longitude, invalidLocationCoordinateDetails, errorCode, errorDescription
}
