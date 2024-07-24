run:
	templ generate
	go build -o course-site
	./course-site lessons
