package installer

import "context"

func RunPipeline(ctx context.Context, modules []Module, opts *InstallOptions, progress func(ProgressEvent)) []Result {
	results := make([]Result, 0, len(modules))
	for _, m := range modules {
		if ctx.Err() != nil {
			results = append(results, Result{
				ModuleName: m.Name(),
				Success:    false,
				Err:        ctx.Err(),
			})
			continue
		}
		r := m.Install(ctx, opts, progress)
		results = append(results, r)
	}
	return results
}

func HasErrors(results []Result) bool {
	for _, r := range results {
		if r.Err != nil {
			return true
		}
	}
	return false
}

func AllPaths(results []Result) []string {
	var paths []string
	for _, r := range results {
		paths = append(paths, r.PathsWritten...)
	}
	return paths
}

func AllWarnings(results []Result) []string {
	var warnings []string
	for _, r := range results {
		warnings = append(warnings, r.Warnings...)
	}
	return warnings
}
