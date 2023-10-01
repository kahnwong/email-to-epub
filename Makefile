run:
	rm -rf emails || exit 0
	rm output.epub || exit 0
	./email-to-epub
