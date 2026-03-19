license-check:
	cd ../backend && go tool addlicense -check -f ../.license_header.txt -ignore "**/build/**" -ignore "**/.gradle/**" -ignore ../android/app/capacitor.build.gradle -ignore ../android/capacitor.settings.gradle ../android

license-fix:
	cd ../backend && go tool addlicense -f ../.license_header.txt -ignore "**/build/**" -ignore "**/.gradle/**" -ignore ../android/app/capacitor.build.gradle -ignore ../android/capacitor.settings.gradle ../android

fix: license-fix

check: license-check

pre-commit: fix check

.PHONY: license-check license-fix fix check
