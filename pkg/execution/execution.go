package execution

import (
	"context"
	"errors"
	"fmt"
	"github.com/masinger/incredible/pkg/execution/customizer"
	"github.com/masinger/incredible/pkg/provider"
	"github.com/masinger/incredible/pkg/provider/source"
	"github.com/masinger/incredible/pkg/specs"
	"go.uber.org/zap"
)

type Execution struct {
	initializedProviders []initializedProvider
}

func NewExecution(
	providers []provider.Provider,
	opts Options,
) (*Execution, error) {
	result := Execution{}

	ctx := opts.Context
	if ctx == nil {
		ctx = context.TODO()
	}

	log := opts.Log
	if log == nil {
		l, err := zap.NewProduction()
		if err != nil {
			return nil, err
		}
		log = l.Sugar()
	}

	for _, currentProvider := range providers {
		name := currentProvider.Name()

		providerLogger := log.WithOptions(zap.Fields(zap.String("provider", name)))
		providerLogger.Debug("Initializing provider")
		providerRuntime := &provider.Runtime{
			Log: providerLogger,
		}

		err := currentProvider.Initialize(ctx, providerRuntime)
		if err == nil {
			result.initializedProviders = append(result.initializedProviders, initializedProvider{
				provider: currentProvider,
				runtime:  providerRuntime,
			})
		} else if errors.Is(err, provider.ErrProviderUnavailable) {
			if opts.FailUnavailableProviders {
				return nil, err
			} else if opts.WarnUnavailableProviders {
				providerLogger.Warnf("%v", err)
			} else {
				providerLogger.Debugf("%v", err)
			}
		} else {
			return nil, err
		}
	}

	return &result, nil
}

func (e Execution) LoadSources(
	ctx context.Context,
	manifest *specs.Manifest,
) (customizer.Customizer, error) {
	providersToBeUsed := map[initializedProvider]bool{}
	srcProviders := map[specs.Source]initializedProvider{}
	for _, asset := range manifest.Assets {
		foundProvider := false
		for _, currentProvider := range e.initializedProviders {
			if currentProvider.provider.SupportsSource(&asset.Src) {
				providersToBeUsed[currentProvider] = true
				srcProviders[asset.Src] = currentProvider
				foundProvider = true
				break
			}
		}
		if !foundProvider {
			return nil, fmt.Errorf("no provider for asset") // TODO: which asset?
		}
	}

	for providerToBeUsed := range providersToBeUsed {
		err := providerToBeUsed.provider.PrepareUsage(ctx)
		if err != nil {
			return nil, err
		}
	}

	var customizers customizer.Customizers

	for _, asset := range manifest.Assets {
		currentAsset := asset
		srcProvider := srcProviders[asset.Src]
		loadedSrc, err := srcProvider.provider.Open(ctx, &currentAsset.Src)
		if err != nil {
			return nil, err
		}
		binarySource, isBinarySource := loadedSrc.(source.BinarySource)
		if isBinarySource {
			var path string
			customizers = append(
				customizers,
				customizer.TempFile(ctx, binarySource, &path),
			)
			for _, mapping := range asset.Mappings {
				currentMapping := mapping
				if currentMapping.Env != nil {
					customizers = append(
						customizers,
						customizer.EnvValue(
							currentMapping.Env.Name,
							&path,
						),
					)
				}
			}
		} else if valueSource, isValueSource := loadedSrc.(source.ValueSource); isValueSource {
			var value string
			customizers = append(
				customizers,
				customizer.LoadValue(ctx, valueSource, &value),
			)
			for _, mapping := range asset.Mappings {
				currentMapping := mapping
				if currentMapping.Env != nil {
					customizers = append(
						customizers,
						customizer.EnvValue(
							mapping.Env.Name,
							&value,
						),
					)
				}
			}
		}
	}

	return customizers.Apply, nil
}
