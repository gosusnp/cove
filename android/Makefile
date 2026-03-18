license-check:
	cd ../backend && go tool addlicense -check -f ../.license_header.txt -ignore "**/build/**" -ignore "**/.gradle/**" ../android

license-fix:
	cd ../backend && go tool addlicense -f ../.license_header.txt -ignore "**/build/**" -ignore "**/.gradle/**" ../android

fix: license-fix

check: license-check

.PHONY: license-check license-fix fix check
