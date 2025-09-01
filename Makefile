.PHONY: help serve serve-drafts build clean new-post new-post-es submodules

# Default target
help:
	@echo "Available commands:"
	@echo "  serve        - Start Hugo development server"
	@echo "  serve-drafts - Start Hugo server with drafts and future posts"
	@echo "  build        - Build site for production"
	@echo "  clean        - Clean build artifacts"
	@echo "  new-post     - Create new English blog post (usage: make new-post TITLE='My Post Title')"
	@echo "  new-post-es  - Create new Spanish blog post (usage: make new-post-es TITLE='Mi Título')"
	@echo "  submodules   - Initialize and update git submodules"

# Start development server
serve:
	hugo server --gc --noHTTPCache --navigateToChanged --disableFastRender

# Start server with drafts and future posts
serve-drafts:
	hugo server -D -F --gc --noHTTPCache --navigateToChanged --disableFastRender

# Build for production
build:
	hugo --gc --minify

# Clean build artifacts
clean:
	rm -rf public/
	rm -rf resources/

# Create new English blog post with page bundle
new-post:
	@if [ -z "$(TITLE)" ]; then \
		echo "Error: TITLE is required. Usage: make new-post TITLE='My Post Title'"; \
		exit 1; \
	fi
	@SLUG=$$(echo "$(TITLE)" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-\|-$$//g'); \
	mkdir -p "content/posts/$$SLUG"; \
	hugo new "posts/$$SLUG/index.md"; \
	echo "Created page bundle: content/posts/$$SLUG/"; \
	echo "Main file: content/posts/$$SLUG/index.md"; \
	echo "You can now add images directly to the content/posts/$$SLUG/ folder"

# Create new Spanish blog post with page bundle
new-post-es:
	@if [ -z "$(TITLE)" ]; then \
		echo "Error: TITLE is required. Usage: make new-post-es TITLE='Mi Título'"; \
		exit 1; \
	fi
	@SLUG=$$(echo "$(TITLE)" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-\|-$$//g'); \
	mkdir -p "content/posts/$$SLUG"; \
	hugo new "posts/$$SLUG/index.es.md"; \
	echo "Created Spanish page bundle: content/posts/$$SLUG/"; \
	echo "Main file: content/posts/$$SLUG/index.es.md"; \
	echo "You can now add images directly to the content/posts/$$SLUG/ folder"

# Initialize and update submodules
submodules:
	git submodule update --init --recursive