# Log: test-runner

- Filter by argv (not env): visible, does not leak into nested children.
- Golden dir baked as absolute path through Program.GoldenDir (set in driver.go) to avoid touching modules_check.go, which the other agent edits.
- Runner C kept in test_runner.go as a string; did not need to move runtime_test.h into every test build.
- Test blocks already accept io (effects of test bodies are discarded). Verified with Test.golden and Test.panics samples.
- Fixed std_test.go and modules_test.go expectations for the new exit code and summary.
