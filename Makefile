run:
	rm -rf emails || exit 0
	rm output.epub || exit 0
	./build/email-to-epub
	./upload.sh
