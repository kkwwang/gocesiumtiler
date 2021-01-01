package std_algorithm_manager

import (
	"github.com/mfbonfigli/gocesiumtiler/internal/converters"
	"github.com/mfbonfigli/gocesiumtiler/internal/converters/coordinate/proj4_coordinate_converter"
	"github.com/mfbonfigli/gocesiumtiler/internal/converters/elevation/geoid_elevation_corrector"
	"github.com/mfbonfigli/gocesiumtiler/internal/converters/elevation/offset_elevation_corrector"
	"github.com/mfbonfigli/gocesiumtiler/internal/converters/elevation/pipeline_elevation_corrector"
	"github.com/mfbonfigli/gocesiumtiler/internal/converters/geoid_offset/gh_offset_calculator"
	"github.com/mfbonfigli/gocesiumtiler/internal/octree"
	"github.com/mfbonfigli/gocesiumtiler/internal/octree/random_trees"
	"github.com/mfbonfigli/gocesiumtiler/internal/tiler"
	"github.com/mfbonfigli/gocesiumtiler/pkg/algorithm_manager"
	"log"
)

type StandardAlgorithmManager struct {
	options             *tiler.TilerOptions
	coordinateConverter converters.CoordinateConverter
	elevationCorrector  converters.ElevationCorrector
	treeStructure       octree.ITree
}

func NewAlgorithmManager(opts *tiler.TilerOptions) algorithm_manager.AlgorithmManager {
	coordinateConverter := proj4_coordinate_converter.NewProj4CoordinateConverter()
	ellipsoidToGeoidOffsetCalculator := gh_offset_calculator.NewEllipsoidToGeoidGHOffsetCalculator(coordinateConverter)
	elevationCorrectionAlgorithm := evaluateElevationCorrectionAlgorithm(opts, ellipsoidToGeoidOffsetCalculator, coordinateConverter)

	algorithmManager := &StandardAlgorithmManager{
		options:             opts,
		coordinateConverter: coordinateConverter,
		elevationCorrector:  elevationCorrectionAlgorithm,
		treeStructure:       evaluateTreeAlgorithm(opts, coordinateConverter, elevationCorrectionAlgorithm),
	}

	return algorithmManager
}

func (algorithmManager *StandardAlgorithmManager) GetElevationCorrectionAlgorithm() converters.ElevationCorrector {
	return algorithmManager.elevationCorrector
}

func (algorithmManager *StandardAlgorithmManager) GetTreeAlgorithm() octree.ITree {
	return algorithmManager.treeStructure
}

func (algorithmManager *StandardAlgorithmManager) GetCoordinateConverterAlgorithm() converters.CoordinateConverter {
	return algorithmManager.coordinateConverter
}

func evaluateElevationCorrectionAlgorithm(options *tiler.TilerOptions, ellipsoidToGeoidOffsetCalculator converters.EllipsoidToGeoidOffsetCalculator, converter converters.CoordinateConverter) converters.ElevationCorrector {
	var elevationCorrectors []converters.ElevationCorrector
	{
		offset_elevation_corrector.NewOffsetElevationCorrector(options.ZOffset)
	}

	if options.EnableGeoidZCorrection {
		elevationCorrectors = append(elevationCorrectors, geoid_elevation_corrector.NewGeoidElevationCorrector(options.Srid, ellipsoidToGeoidOffsetCalculator))
	}

	return pipeline_elevation_corrector.NewPipelineElevationCorrector(elevationCorrectors)
}

func evaluateTreeAlgorithm(options *tiler.TilerOptions, converter converters.CoordinateConverter, elevationCorrection converters.ElevationCorrector) octree.ITree {
	switch options.Strategy {
	case tiler.BoxedRandom:
		return random_trees.NewBoxedRandomTree(options, converter, elevationCorrection)
	case tiler.FullyRandom:
		return random_trees.NewRandomTree(options, converter, elevationCorrection)
	}

	log.Fatal("Unrecognized strategy")
	return nil
}
