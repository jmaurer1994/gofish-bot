package obs

import (
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"time"
)

type GoobsClient struct {
	server string
	client *goobs.Client
}

func NewGoobsClient(_server, password string) (*GoobsClient, error) {
	gc := &GoobsClient{
		server: _server,
	}

	client, err := goobs.New(_server, goobs.WithPassword(password))

	if err != nil {
		return nil, err
	}

	gc.client = client

	return gc, nil
}

func (gc *GoobsClient) ToggleSourceVisibility(sceneName, itemName string) error {
	params := sceneitems.NewGetSceneItemListParams().WithSceneName(sceneName)

	sceneList, err := gc.client.SceneItems.GetSceneItemList(params)
	if err != nil {
        return fmt.Errorf("Error geting scenelist: %s", err)
	}

	// find the ID of our source, while hiding all others
	var sourceID int
	for _, item := range sceneList.SceneItems {
		if item.SourceName == itemName {
			sourceID = item.SceneItemID
		}
	}

	// then show our source
	err = gc.setSourceVisibility(sceneName, sourceID, false)

    if(err != nil){
        return err
    }

	time.Sleep(3 * time.Second)

	err = gc.setSourceVisibility(sceneName, sourceID, true)

    if(err != nil){
        return err
    }

	return nil
}

func (gc *GoobsClient) setSourceVisibility(scene string, sourceID int, visible bool) error {
	params := &sceneitems.SetSceneItemEnabledParams{
		SceneName:        &scene,
		SceneItemId:      &sourceID,
		SceneItemEnabled: &visible,
	}
	_, err := gc.client.SceneItems.SetSceneItemEnabled(params)
	if err != nil {
		return err
	}

    return nil
}
