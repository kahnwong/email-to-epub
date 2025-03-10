run:
	rm -rf emails || exit 0
	rm output*.epub || exit 0
	./main
# 	mv output*.epub "/Users/kahnwong/ereader/newsletters/" # mac
	mv output*.epub "/mnt/hdd/Media/Ereader/newsletters/" # linux
	curl -d "Done" https://ntfy.karnwong.me/email-to-epub
