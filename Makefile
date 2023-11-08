generate-base-profile:
	@curl -o profiles/base.pprof http://localhost:8080/debug/pprof/trace?seconds=30

generate-result-profile:
	@curl -o profiles/result.pprof http://localhost:8080/debug/pprof/trace?seconds=30