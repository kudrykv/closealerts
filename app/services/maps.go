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

	"go.uber.org/zap"
	"golang.org/x/sync/singleflight"
	"gorm.io/gorm"
)

type Maps struct {
	sf    *singleflight.Group
	mapz  repositories.Maps
	log   *zap.SugaredLogger
	alert Alerts
}

func NewMaps(
	log *zap.SugaredLogger,
	mapz repositories.Maps,
	alert Alerts,
) Maps {
	return Maps{
		log:   log,
		sf:    &singleflight.Group{},
		mapz:  mapz,
		alert: alert,
	}
}

func (r Maps) Get(ctx context.Context, alerts types2.Alerts) (bool, types2.Map, []byte, error) {
	areas := alerts.Areas().Sort()
	alertsKey := areas.Join(",")

	mapz, err := r.mapz.Get(ctx, alertsKey)
	if err == nil {
		return true, mapz, nil, nil
	}

	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, mapz, nil, fmt.Errorf("get map: %w", err)
	}

	r.log.Infow("no map for given alerts set yet", "areas", areas)

	hash := md5.New()
	hash.Write([]byte(alertsKey))
	out := hash.Sum(nil)
	filename := fmt.Sprintf("%x", out)

	if _, err := exec.LookPath("convert"); err != nil {
		return false, mapz, nil, fmt.Errorf("look path convert: %w", err)
	}

	bts, err := r.alert.GetMapSVGBytes(ctx)
	if err != nil {
		return false, mapz, nil, fmt.Errorf("get map svg bytes: %w", err)
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
