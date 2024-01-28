package obs

import (
	"fmt"
	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/sources"
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

func (gc *GoobsClient) ToggleSourceVisibility(sceneName, sourceName string) error {

	// find the ID of our source, while hiding all others
	sourceId, err := gc.getSourceIdByName(sceneName, sourceName)
	if err != nil {
		return err
	}

	// hide source
	err = gc.setSourceVisibility(sceneName, sourceId, false)

	if err != nil {
		return err
	}

	time.Sleep(4 * time.Second)

	// show source
	err = gc.setSourceVisibility(sceneName, sourceId, true)

	if err != nil {
		return err
	}

	return nil
}

func (gc *GoobsClient) getSourceIdByName(sceneName, sourceName string) (int, error) {
	params := sceneitems.NewGetSceneItemListParams().WithSceneName(sceneName)

	sceneList, err := gc.client.SceneItems.GetSceneItemList(params)
	if err != nil {
		return -1, fmt.Errorf("Error getting scenelist: %s", err)
	}

	for _, item := range sceneList.SceneItems {
		if item.SourceName == sourceName {
			return item.SceneItemID, nil
		}
	}

	return -1, fmt.Errorf("Source not found in scene item list")
}

func (gc *GoobsClient) setSourceVisibility(scene string, sourceId int, visible bool) error {
	params := &sceneitems.SetSceneItemEnabledParams{
		SceneName:        &scene,
		SceneItemId:      &sourceId,
		SceneItemEnabled: &visible,
	}
	_, err := gc.client.SceneItems.SetSceneItemEnabled(params)

	return err
}

// TODO: use env var for base screenshot location?
func (gc *GoobsClient) ScreenshotSource(sourceName string) error {
	params := sources.NewSaveSourceScreenshotParams().WithSourceName(sourceName).WithImageCompressionQuality(100).WithImageFilePath(fmt.Sprintf("M:\\screenshots\\%s", sourceName))
	_, err := gc.client.Sources.SaveSourceScreenshot(params)

	return err
}
