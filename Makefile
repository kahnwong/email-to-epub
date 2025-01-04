run:
	rm -rf emails || exit 0
	rm output*.epub || exit 0
	./main
	mv output*.epub "/Users/kahnwong/ereader/newsletters/"
