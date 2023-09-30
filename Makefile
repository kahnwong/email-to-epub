run:
	rm -rf emails || exit 0
	rm output.epub || exit 0
	./build/email-to-epub
	email-to-epub emails/*.eml
	./upload.sh
