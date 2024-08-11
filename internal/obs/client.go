package obs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/andreykaipov/goobs"
	"github.com/andreykaipov/goobs/api/requests/sceneitems"
	"github.com/andreykaipov/goobs/api/requests/sources"
	"github.com/minio/minio-go/v7"
)

type GoobsClient struct {
	server              string
	client              *goobs.Client
	screenshotDirectory string
	screenshotFormat    string
	screenshotQuality   float64
}

func NewGoobsClient(server, password, screenshotDirectory, screenshotFormat string, screenshotQuality float64) (*GoobsClient, error) {
	gc := &GoobsClient{
		server:              server,
		screenshotDirectory: screenshotDirectory,
		screenshotFormat:    screenshotFormat,
		screenshotQuality:   screenshotQuality,
	}

	client, err := goobs.New(server, goobs.WithPassword(password))

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

func (gc *GoobsClient) ScreenshotSource(sourceName string) (string, error) {
	params := &sources.GetSourceScreenshotParams{
		SourceName:              &sourceName,
		ImageFormat:             &gc.screenshotFormat,
		ImageCompressionQuality: &gc.screenshotQuality,
	}
	screenshot, err := gc.client.Sources.GetSourceScreenshot(params)

	if err != nil {
		return "", err
	}

	data := screenshot.ImageData[strings.IndexByte(screenshot.ImageData, ',')+1:]
	return data, nil
}

func (gc *GoobsClient) ScreenhotToBucket(sourceName, fileName, bucket string, s3conn *minio.Client) error {

	screenshot, err := gc.ScreenshotSource(sourceName)

	if err != nil {
		return errors.Join(errors.New("Error screenshotting souce"), err)
	}

	reader := base64.NewDecoder(base64.StdEncoding, strings.NewReader(screenshot))

	_, err = s3conn.PutObject(
		context.Background(),
		bucket,
		fmt.Sprintf("%s.png", fileName),
		reader,
		-1,
		minio.PutObjectOptions{
			ContentType: "image/png",
		},
	)

	if err != nil {
		return errors.Join(errors.New("Error saving object to storage"), err)
	}

	return nil
}
