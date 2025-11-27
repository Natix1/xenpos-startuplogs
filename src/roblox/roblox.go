package roblox

import (
	"encoding/json"
	"fmt"

	"github.com/natix1/xenpos-startuplogs/src/rest"
)

func GetUniverseIdFromPlaceId(placeId int) (int, error) {
	path := fmt.Sprintf("universes/v1/places/%d/universe", placeId)
	body, err := rest.RobloxGet(path)
	if err != nil {
		return 0, err
	}

	var response UniverseIdResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	return response.UniverseId, nil
}

func GetUniverse(universeId int) (*Universe, error) {
	path := fmt.Sprintf("cloud/v2/universes/%d", universeId)
	body, err := rest.RobloxGet(path)
	if err != nil {
		return nil, err
	}

	var universe Universe
	err = json.Unmarshal(body, &universe)
	if err != nil {
		return nil, err
	}

	return &universe, nil
}

func GetUser(userId int) (*User, error) {
	path := fmt.Sprintf("cloud/v2/users/%d", userId)
	body, err := rest.RobloxGet(path)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
