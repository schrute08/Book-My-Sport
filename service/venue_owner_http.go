package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/amancooks08/BookMySport/domain"
)

// Add a  venue
func AddVenue(CustomerServices Services) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var venue domain.Venue
		err := json.NewDecoder(req.Body).Decode(&venue)
		if err != nil {
			http.Error(rw, "invalid request payload", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		err = CustomerServices.CheckVenue(req.Context(), venue.Name, venue.Contact, venue.Email)
		if err != nil {
			msg := domain.Message{
				Message: fmt.Sprintf("%s", err),
			}
			json_response, _ := json.Marshal(msg)

			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(json_response)
			return
		}

		if venue.Name == "" || venue.Address == "" || venue.City == "" || venue.State == "" || len(venue.Games) == 0 {
			msg := domain.Message{
				Message: "please don't leave any field empty",
			}
			json_response, _ := json.Marshal(msg)

			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(json_response)
			return
		}
		userID := GetUserID(req, rw)
		venue.OwnerID = userID
		if validateContact(venue.Contact) && validateEmail(venue.Email) {
			err := CustomerServices.AddVenue(req.Context(), venue)
			if err != nil {
				msg := domain.Message{
					Message: "failed to add venue",
				}

				json_response, _ := json.Marshal(msg)
				rw.Header().Add("Content-Type", "application/json")
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write(json_response)
				return
			}
		} else {
			msg := domain.Message{
				Message: "invalid email or contact",
			}

			json_response, _ := json.Marshal(msg)
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(json_response)
			return
		}

		// Write the response
		response := domain.Venue{
			ID:      venue.ID,
			Name:    venue.Name,
			Address: venue.Address,
			City:    venue.City,
			State:   venue.State,
			Contact: venue.Contact,
			Email:   venue.Email,
			Opening: venue.Opening,
			Closing: venue.Closing,
			Price:   venue.Price,
			Games:   venue.Games,
			Rating:  venue.Rating,
			OwnerID: venue.OwnerID,
		}
		json_response, err := json.Marshal(response)
		if err != nil {
			http.Error(rw, "failed to marshal response", http.StatusInternalServerError)
			return
		}

		rw.Header().Add("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		rw.Write(json_response)

	})
}

// Update a venue
func UpdateVenue(CustomerServices Services) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodPut {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var venue domain.Venue
		err := json.NewDecoder(req.Body).Decode(&venue)
		if err != nil {
			http.Error(rw, "invalid request payload", http.StatusBadRequest)
			return
		}
		defer req.Body.Close()

		//Get the userID and venueID from the jwt token and URL respectively

		userID:= GetUserID(req, rw)
		venueID := GetVenueID(req)
		if userID == 0 || venueID == 0 {
			msg := domain.Message{
				Message: "invalid user or venue ID",
			}
			json_response, _ := json.Marshal(msg)
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write(json_response)
			return
		}

		if venue.Name == "" || venue.Address == "" || venue.City == "" || venue.State == "" {
			http.Error(rw, "invalid request payload", http.StatusBadRequest)
			return
		}

		if validateContact(venue.Contact) && validateEmail(venue.Email) {

			err := CustomerServices.UpdateVenue(req.Context(), venue, userID, venueID)
			if err != nil {
				http.Error(rw, "error: updating venue", http.StatusInternalServerError)
				return
			}
			if err == nil {
				responseMessage := "venue updated successfully"
				rw.Header().Add("Content-Type", "application/json")
				rw.WriteHeader(http.StatusAccepted)
				rw.Write([]byte(responseMessage))
			}
		} else {
			http.Error(rw, "invalid email or contact information.", http.StatusBadRequest)
			return
		}
	})
}

// Check availbility at a venue
func CheckAvailability(CustomerServices Services) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodGet {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		// vars := mux.Vars(req)
		// venueID, err := strconv.Atoi(vars["venue_id"])
		// if err != nil {
		// 	http.Error(rw, fmt.Sprint(err)+": Invalid ID", http.StatusBadRequest)
		// 	return
		// }

		venueID := GetVenueID(req)

		// Check if date is present or not
		if req.URL.Query().Get("date") == "" {
			http.Error(rw, "please enter date if not entered or correct it if not added properly.", http.StatusBadRequest)
			return
		}
		date, err := time.Parse("2006-01-02", req.URL.Query().Get("date"))
		if err != nil {
			http.Error(rw, "invalid date format", http.StatusBadRequest)
			return
		}
		if date.After(time.Now().Truncate(24*time.Hour)) || date.Equal(time.Now().Truncate(24*time.Hour)) {
			availabileSlots, err := CustomerServices.CheckAvailability(req.Context(), venueID, date.Format("2006-01-02"))
			if err != nil {
				http.Error(rw, fmt.Sprintf("%s", err), http.StatusInternalServerError)
				return
			}

			respBytes, err := json.Marshal(availabileSlots)
			if err != nil {
				http.Error(rw, "failed to marshal response", http.StatusInternalServerError)
				return
			}

			rw.Header().Add("Content-Type", "application/json")
			rw.Write(respBytes)
		} else {
			http.Error(rw, "invalid date - please selct an upcoming date.", http.StatusBadRequest)
			return
		}
	})
}

// Delete a venue
func DeleteVenue(CustomerServices Services) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		if req.Method != http.MethodDelete {
			rw.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		userID := GetUserID(req, rw)
		venueID := GetVenueID(req)
		//Check if "venue_id" key is not found in vars

		if userID == 0 || venueID == 0 {
			fmt.Println(userID, venueID)
			msg := domain.Message{
				Message: "invalid user or venue ID",
			}
			json_response, _ := json.Marshal(msg)
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusForbidden)
			rw.Write(json_response)
			return
		}

		err := CustomerServices.DeleteVenue(req.Context(), userID, venueID)
		if err != nil && err.Error() == "you are not the owner of this venue" {
			msg := domain.Message{
				Message: fmt.Sprintf("%s", err),
			}
			jsonResponse, _ := json.Marshal(msg)
			rw.Header().Add("Content-Type", "application/json")
			rw.WriteHeader(http.StatusForbidden)
			rw.Write(jsonResponse)
			return
		} else if err != nil {
			http.Error(rw, "error: deleting venue", http.StatusInternalServerError)
			return
		}

		resp := domain.Message{
			Message: "venue deleted successfully",
		}
		respBytes, err := json.Marshal(resp)
		if err != nil {
			http.Error(rw, "failed to marshal response", http.StatusInternalServerError)
			return
		}

		rw.Header().Add("Content-Type", "application/json")
		rw.Write(respBytes)
	})
}
