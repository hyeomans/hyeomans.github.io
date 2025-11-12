.PHONY: new-post help

help:
	@echo "Available targets:"
	@echo "  make new-post    - Create a new blog post"
	@echo "  make help        - Show this help message"

new-post:
	@echo "Create a new blog post"
	@echo ""
	@echo "Select language:"
	@echo "  1) English"
	@echo "  2) Spanish"
	@read -p "Enter choice [1-2]: " lang_choice; \
	if [ "$$lang_choice" = "1" ]; then \
		lang="en"; \
		lang_name="English"; \
	elif [ "$$lang_choice" = "2" ]; then \
		lang="es"; \
		lang_name="Spanish"; \
	else \
		echo "Invalid choice. Exiting."; \
		exit 1; \
	fi; \
	echo ""; \
	read -p "Enter post title: " title; \
	if [ -z "$$title" ]; then \
		echo "Title cannot be empty. Exiting."; \
		exit 1; \
	fi; \
	echo ""; \
	read -p "Enter post description: " description; \
	if [ -z "$$description" ]; then \
		echo "Description cannot be empty. Exiting."; \
		exit 1; \
	fi; \
	echo ""; \
	read -p "Enter tags (comma-separated, e.g., golang,typescript,programming): " tags; \
	if [ -z "$$tags" ]; then \
		echo "Tags cannot be empty. Exiting."; \
		exit 1; \
	fi; \
	slug=$$(echo "$$title" | tr '[:upper:]' '[:lower:]' | sed 's/[^a-z0-9]/-/g' | sed 's/--*/-/g' | sed 's/^-//;s/-$$//'); \
	post_dir="src/content/posts/$$lang/$$slug"; \
	mkdir -p "$$post_dir"; \
	post_file="$$post_dir/index.md"; \
	hero_file="$$post_dir/hero.svg"; \
	cp public/placehold.svg "$$hero_file"; \
	today=$$(date -u +"%Y-%m-%dT%H:%M:%S%z" | sed 's/\(..\)$$/:\1/'); \
	tags_formatted=$$(echo "$$tags" | sed 's/,/", "/g'); \
	echo "---" > "$$post_file"; \
	echo "title: \"$$title\"" >> "$$post_file"; \
	echo "description: \"$$description\"" >> "$$post_file"; \
	echo "pubDate: $$today" >> "$$post_file"; \
	echo "author: \"Hector Yeomans\"" >> "$$post_file"; \
	echo "tags: [\"$$tags_formatted\"]" >> "$$post_file"; \
	echo "lang: \"$$lang\"" >> "$$post_file"; \
	echo "draft: false" >> "$$post_file"; \
	echo "heroImage: \"./hero.svg\"" >> "$$post_file"; \
	echo "heroAlt: \"Hero image for $$title\"" >> "$$post_file"; \
	echo "---" >> "$$post_file"; \
	echo "" >> "$$post_file"; \
	echo "Write your post content here..." >> "$$post_file"; \
	echo "" >> "$$post_file"; \
	echo "✅ Post created successfully!"; \
	echo "📁 Location: $$post_file"; \
	echo "🌐 Language: $$lang_name"; \
	echo "📝 Title: $$title"; \
	echo ""; \
	echo "You can now edit the post at: $$post_file"
