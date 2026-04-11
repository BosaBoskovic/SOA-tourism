package com.example.blog.service;

import com.example.blog.model.Blog;
import com.example.blog.model.Comment;
import com.example.blog.repository.BlogRepository;
import com.vladsch.flexmark.html.HtmlRenderer;
import com.vladsch.flexmark.parser.Parser;
import com.vladsch.flexmark.util.ast.Node;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Service
@RequiredArgsConstructor
public class BlogService{

    private final BlogRepository blogRepository;

    private final Parser markdownParser = Parser.builder().build();
    private final HtmlRenderer htmlRenderer = HtmlRenderer.builder().build();

    //kreiranje bloga
    public Blog createBlog(String title, String descriptionMarkdown, List<String>imageUrls, String authorUsername){
        Blog blog = new Blog();
        blog.setTitle(title);
        blog.setDescriptionMarkdown(descriptionMarkdown);
        blog.setImageUrls(imageUrls != null ? imageUrls : List.of());
        blog.setAuthorUsername(authorUsername);
        return blogRepository.save(blog);
    }

    public List<Blog> getAllBlogs(){
        return blogRepository.findAll();
    }

    public Optional<Blog> getBlogById(String id){
        return blogRepository.findById(id);
    }

    //pomocna metoda - konvertuje markdown u HTML
    public String renderMarkdown(String markdown){
        if(markdown == null) return "";
        Node document = markdownParser.parse(markdown);
        return htmlRenderer.render(document);
    }

    public Blog addComment(String blogId, String authorUsername, String text){
        Blog blog = blogRepository.findById(blogId).orElseThrow(()-> new RuntimeException("Blog nije pronadjen"));

        Comment comment = new Comment();
        comment.setAuthorUsername(authorUsername);
        comment.setText(text);

        blog.getComments().add(comment);
        return blogRepository.save(blog);
    }

    public Blog editComment(String blogId, String commentId, String username, String newText){
        Blog blog = blogRepository.findById(blogId).orElseThrow(()-> new RuntimeException("Blog nije pronadjen"));

        Comment comment = blog.getComments().stream().filter(c -> c.getId().equals(commentId)).findFirst().orElseThrow(()-> new RuntimeException("Komentar nije pronadjen"));

        if(!comment.getAuthorUsername().equals(username)){
            throw new RuntimeException("Nije tvoj komentar");
        }

        comment.setText(newText);
        comment.setLastModifiedAt(LocalDateTime.now());
        return blogRepository.save(blog);
    }

    public Blog likeBlog(String blogId, String username) {
        Blog blog = blogRepository.findById(blogId).orElseThrow(() -> new RuntimeException("Blog nije pronadjen"));

        if (blog.getLikes().contains(username)) {
            blog.getLikes().remove(username);
        } else {
            blog.getLikes().add(username);
        }

        return blogRepository.save(blog);
    }
}