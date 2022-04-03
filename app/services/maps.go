package services

import (
	"closealerts/app/repositories"
	types2 "closealerts/app/repositories/types"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"regexp"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type Maps struct {
	mapz  repositories.Maps
	log   *zap.SugaredLogger
	alert Alerts
}

func NewMaps(
	log *zap.SugaredLogger,
	mapz repositories.Maps,
) Maps {
	return Maps{
		log:  log,
		mapz: mapz,
	}
}

func (r Maps) Get(ctx context.Context, alerts types2.Alerts) (bool, types2.Map, []byte, error) {
	areas := alerts.Areas().Sort()
	alertsKey := areas.Join(",")

	mapz, ok, err := r.Exists(ctx, alerts)
	if err != nil {
		return false, mapz, nil, fmt.Errorf("exists: %w", err)
	}

	if ok {
		return true, mapz, nil, nil
	}

	r.log.Infow("no map for given alerts set yet", "areas", areas)

	hash := md5.New()
	hash.Write([]byte(alertsKey))
	out := hash.Sum(nil)
	filename := fmt.Sprintf("%x", out)

	if _, err := exec.LookPath("convert"); err != nil {
		return false, mapz, nil, fmt.Errorf("look path convert: %w", err)
	}

	bts, err := os.ReadFile("map.svg")
	if err != nil {
		return false, mapz, nil, fmt.Errorf("os read file map.svg: %w", err)
	}

	if bts, err = r.Paint(bts, alerts); err != nil {
		return false, mapz, nil, fmt.Errorf("paint: %w", err)
	}

	filenameSVG := filename + ".svg"
	filenamePNG := filename + ".png"

	if err := os.WriteFile(filenameSVG, bts, 0700); err != nil {
		return false, mapz, nil, fmt.Errorf("os write file: %w", err)
	}

	if err := exec.CommandContext(ctx, "convert", "-resize", "1500x", filenameSVG, filenamePNG).Run(); err != nil {
		return false, mapz, nil, fmt.Errorf("command run: convert -resize 1500x %s %s: %w", filenameSVG, filenamePNG, err)
	}

	bts, err = os.ReadFile(filenamePNG)
	if err != nil {
		return false, mapz, nil, fmt.Errorf("os read file %s: %w", filenamePNG, err)
	}

	if err := os.Remove(filenameSVG); err != nil {
		return false, mapz, nil, fmt.Errorf("remove %s: %w", filenameSVG, err)
	}

	if err := os.Remove(filenamePNG); err != nil {
		return false, mapz, nil, fmt.Errorf("remove %s: %w", filenamePNG, err)
	}

	return false, mapz, bts, nil
}

func (r Maps) Save(ctx context.Context, alerts types2.Alerts, fileID string) (types2.Map, error) {
	areas := alerts.Areas().Sort()
	alertsKey := areas.Join(",")

	mapz, err := r.mapz.Save(ctx, alertsKey, fileID)
	if err != nil {
		return mapz, fmt.Errorf("save: %w", err)
	}

	return mapz, nil
}

func (r Maps) Exists(ctx context.Context, alerts types2.Alerts) (types2.Map, bool, error) {
	areas := alerts.Areas().Sort()
	alertsKey := areas.Join(",")

	mapz, err := r.mapz.Get(ctx, alertsKey)
	if err == nil {
		return mapz, true, nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return mapz, false, nil
	}

	return mapz, false, fmt.Errorf("mapz get: %w", err)
}

func (r Maps) Paint(bts []byte, alerts types2.Alerts) ([]byte, error) {
	if len(alerts) == 0 {
		return bts, nil
	}

	for _, alert := range alerts {
		regex, err := regexp.Compile(`(<[^>]+fill=)"[^"]+"([^>]+data-oblast="` + alert.ID + ")")
		if err != nil {
			return nil, fmt.Errorf("regexp compile: %w", err)
		}

		bts = regex.ReplaceAll(bts, []byte(`$1"rgba(230,25,25,1)"$2`))

		regex, err = regexp.Compile(`(<[^>]+data-oblast="` + alert.ID + `[^>]+fill=)"[^"]+"`)
		if err != nil {
			return nil, fmt.Errorf("regexp compile: %w", err)
		}

		bts = regex.ReplaceAll(bts, []byte(`$1"rgba(230,25,25,1)"`))

		regex, err = regexp.Compile(`(<[^>]+fill-opacity=)"[^"]+"([^>]+data-oblast="` + alert.ID + `)`)
		if err != nil {
			return nil, fmt.Errorf("regexp compile: %w", err)
		}

		bts = regex.ReplaceAll(bts, []byte(`$1"0.4"$2`))

		regex, err = regexp.Compile(`(<[^>]+data-oblast="` + alert.ID + `[^>]+fill-opacity=)"[^"]+"`)
		if err != nil {
			return nil, fmt.Errorf("regexp compile: %w", err)
		}

		bts = regex.ReplaceAll(bts, []byte(`$1"0.4"`))
	}

	return bts, nil
}
