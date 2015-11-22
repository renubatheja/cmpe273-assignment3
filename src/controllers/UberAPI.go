package controllers

import (
    "encoding/json"
    "fmt"
    "net/http"
    "io/ioutil"
    "strconv"
    "uber"
    "model"
)

//-------------------------------Structs to Process Uber API----------------------------
type Response struct {
	Results []PricesStruct `json:"prices"`
}

type PricesStruct struct {
	LowEstimate float64 `json:"low_estimate"`
	HighEstimate float64 `json:"high_estimate"`
	Duration float64 `json:"duration"`
	Distance float64 `json:"distance"`
}

type ProductTypeResponse struct {
	Results []ProductTypesStruct `json:"products"`
}

type ProductTypesStruct struct {
	DisplayName string `json:"display_name"`
	ProductID string `json:"product_id"`
}

var Client *uber.Client

/*
* Function to generate and capture OAuth token
*/

func GetOAuth() {
	  	  SERVER_TOKEN := "XuKMSJ2-MqcHPOT2XXKJChZYGcXkyNJgYG2LOSQj"
		  CLIENT_ID := "1LttUcxVT_9BsWf4n__ysFjka7MAm99g"
		  CLIENT_SECRET := "TsdZnMzMUXwraDYP4MOKCoOXPkrZGO2i8mGkmGuA"
		  REDIRECT_URL := "http://localhost:7635/"
		  Client = uber.NewClient(SERVER_TOKEN)
		  err := Client.AutOAuth( CLIENT_ID, CLIENT_SECRET, REDIRECT_URL, "request",)
		  if (err != nil && err.Error() != "EOF") {
		  	fmt.Println("Error received while generating OAuth Access Token")
		  }
}

/*
* Function to make Uber API Price Estimate call to fetch price estimates from starting location to end location
*/
func CallUberAPIForPriceEstimates(start_latitude float64, start_longitude float64, end_latitude float64, end_longitude float64) (float64, float64, float64, bool, string, string){
	var query string
	var invalidPricingDetails bool
	invalidPricingDetails = false
	var errorDescription string
	var errorCode string
	
	errorDescription = ""
	errorCode = ""
	startLatitude := strconv.FormatFloat(start_latitude, 'f', -1, 32)
	startLongitude := strconv.FormatFloat(start_longitude, 'f', -1, 32)
	endLatitude := strconv.FormatFloat(end_latitude, 'f', -1, 32)
	endLongitude := strconv.FormatFloat(end_longitude, 'f', -1, 32)
	query = "https://api.uber.com/v1/estimates/price?start_latitude=" + string(startLatitude) + "&start_longitude=" + string(startLongitude) + "&end_latitude=" + string(endLatitude) + "&end_longitude=" + string(endLongitude) + "&server_token=XuKMSJ2-MqcHPOT2XXKJChZYGcXkyNJgYG2LOSQj"
	resp, err := http.Get(query)
	
	if err != nil {
		invalidPricingDetails = true
		errorCode = "UBER_PRICE_ESTIMATE_API_ERROR"
		errorDescription = "Error received while executing Uber's Price Estimates API call. Please check the location details."
		fmt.Println("Error : Error received while executing Uber's Price Estimates API call! : ", err);
		fmt.Println("------------------------------------------------------------")
		return 0, 0, 0, invalidPricingDetails, errorCode, errorDescription
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		errorCode = "UBER_PRICE_ESTIMATE_API_ERROR"
		errorDescription = "Unable to read response from Uber's Price Estimates API. Please try again later."
		fmt.Println("Error : Error received while reading response from Uber's Price Estimates API call! : ", err);
		fmt.Println("------------------------------------------------------------")	
		invalidPricingDetails = true
		return 0, 0, 0, invalidPricingDetails, errorCode, errorDescription
	}
		
	var response Response
	errUnmarshal := json.Unmarshal(body, &response)
	if errUnmarshal != nil {
		errorCode = "UBER_PRICE_ESTIMATE_API_ERROR"
		errorDescription = "Unable to unmarshall response from Uber's Price Estimates API. Please try again later."
		fmt.Println("Error : Error received while unmashalling response from Uber's Price Estimates API call!: ", errUnmarshal);
		fmt.Println("------------------------------------------------------------")	
		invalidPricingDetails = true
		return 0, 0, 0, invalidPricingDetails, errorCode, errorDescription
	}
		
	var uberCost, minUberCost float64
	var uberDuration float64
	var uberDistance float64
	if(len(response.Results) > 0){
		//find lowest uber cost and then lowest uber duration
		minUberCost = 999999999
		for index := 0; index < len(response.Results); index++ {
	 		uberCost = response.Results[index].HighEstimate
	 		//fmt.Println("uberCost : ",uberCost)
	 		if(uberCost < minUberCost && uberCost != 0) {
	 			minUberCost = uberCost
		 		uberDuration = response.Results[index].Duration
		 		uberDistance = response.Results[index].Distance		
	 		}
		}
		//fmt.Println("Lowest estimate : ",minUberCost)
		//fmt.Println("Duration : ",uberDuration)
		//fmt.Println("Distance : ",uberDistance)		
		errorDescription = ""
	} else {
		errorCode = "INVALID_ADDRESS"
		errorDescription = "No results returned! Please enter a valid location"
		fmt.Println("Error : No results returned! Please enter a valid location")
		fmt.Println("------------------------------------------------------------")
		invalidPricingDetails = true
		return 0, 0, 0, invalidPricingDetails, errorCode, errorDescription
	}
	
	return minUberCost, uberDuration, uberDistance, invalidPricingDetails, errorCode, errorDescription
}



/*
* Function to make Uber API Product type call to get the Product ID
*/
var ProductID string

func CallUberAPIForProductType(start_latitude float64, start_longitude float64) string {
	ProductID = ""
	var query string
	
	startLatitude := strconv.FormatFloat(start_latitude, 'f', -1, 32)
	startLongitude := strconv.FormatFloat(start_longitude, 'f', -1, 32)
	query = "https://api.uber.com/v1/products?latitude=" + string(startLatitude) + "&longitude=" + string(startLongitude) + "&server_token=XuKMSJ2-MqcHPOT2XXKJChZYGcXkyNJgYG2LOSQj"
	resp, err := http.Get(query)
	
	if err != nil {
		fmt.Println("Error : Error received while executing Uber's Price Estimates API call! : ", err);
		fmt.Println("------------------------------------------------------------")
	}
	
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println("Error : Error received while reading response from Uber's Price Estimates API call! : ", err);
		fmt.Println("------------------------------------------------------------")	
	}
		
	var response ProductTypeResponse
	errUnmarshal := json.Unmarshal(body, &response)
	if errUnmarshal != nil {
		fmt.Println("Error : Error received while unmashalling response from Uber's Price Estimates API call!: ", errUnmarshal);
		fmt.Println("------------------------------------------------------------")	
	}
		
	if(len(response.Results) > 0){
		for index := 0; index < len(response.Results); index++ {
			if(response.Results[index].DisplayName == "uberX") {
				ProductID = response.Results[index].ProductID
				break
			} 
		}
	} else {
		fmt.Println("Error : No results returned! Please enter a valid location")
		fmt.Println("------------------------------------------------------------")
	}
	fmt.Println("Product ID : " + ProductID)
	return ProductID
}


/*
* Function to make Uber API Ride Request call
*/
func CallUberAPIForRideRequest(lc TripController, StartingFromLocationId string, DestinationLocationId string) model.RideRequestResponse{
	  	  
 	  startLatitude, startLongitude, _, _, _ := GetLocationCoordinatesFromMongoDB(lc, StartingFromLocationId)
 	  destinationLatitude, destinationLongitude, _, _, _ := GetLocationCoordinatesFromMongoDB(lc, DestinationLocationId)
 	  
	  rideResponse := model.RideRequestResponse{}
	  	  	  
	  fmt.Println("Let's see how much time it would take for driver to reach here!")
	  ProductID := CallUberAPIForProductType(startLatitude, startLongitude)
	  //_, err, responseRide := Client.PostRequest("2832a1f5-cfc0-48bb-ab76-7ea7a62060e7",startLatitude,startLongitude,destinationLatitude,destinationLongitude,"")
	  _, err, responseRide := Client.PostRequest(ProductID,startLatitude,startLongitude,destinationLatitude,destinationLongitude,"")
	  //requestRide, err := client.PostRequest("a1111c8c-c720-46c3-8534-2fcdd730040d",37.355,-122.0,37.38,-122.01,"")
	  if (err != nil && err.Error() != "EOF") {
		fmt.Println("Error Occurred while Posting a ride request: " , err)
	  } else {	  	
	  	err := json.Unmarshal([]byte(responseRide), &rideResponse)
	  	if (err != nil && err.Error() != "EOF") {
	  		fmt.Println("Error here : ", err)
	  	} else {
	  		fmt.Println("Driver's ETA : ",rideResponse.Eta)
	  	}
	  }
	  return rideResponse
}
