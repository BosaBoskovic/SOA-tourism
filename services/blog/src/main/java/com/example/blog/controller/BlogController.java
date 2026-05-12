package com.example.blog.controller;

import com.example.blog.exception.BlogAccessDeniedException;
import com.example.blog.exception.BlogNotFoundException;
import com.example.blog.model.Blog;
import com.example.blog.service.BlogService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/blog")
@RequiredArgsConstructor
public class BlogController{

    private final BlogService blogService;

    //kreiranje bloga
    //header: x-username (salje gateway ili klijent)
    @PostMapping
    public ResponseEntity<Map<String, Object>> createBlog(@RequestBody Map<String, Object> body, @RequestHeader("X-Username") String username){
        
        String title = (String) body.get("title");
        String description = (String) body.get("descriptionMarkdown");
        List<String> images = (List<String>) body.get("imageUrls");

        Blog blog = blogService.createBlog(title, description, images, username);
        return ResponseEntity.status(201).body(toBlogResponse(blog, username));
    }

    //dobavljanje svih blogova
    @GetMapping
    public ResponseEntity<List<Map<String, Object>>> getAllBlogs(@RequestHeader("X-Username") String username){
        List<Map<String, Object>> blogs = blogService.getAllBlogsForUser(username).stream()
                .map(blog -> toBlogResponse(blog, username))
                .collect(java.util.stream.Collectors.toList());
        return ResponseEntity.ok(blogs);
    }

    //dobavljanje jednog bloga (sa rendered markdown)
    @GetMapping("/{id}")
    public ResponseEntity<?> getBlogById(@PathVariable String id, @RequestHeader("X-Username") String username){
        try {
            Blog blog = blogService.getBlogByIdForUser(id, username);
            String renderedHtml = blogService.renderMarkdown(blog.getDescriptionMarkdown());
            Map<String, Object> response = toBlogResponse(blog, username);
            response.put("descriptionHtml", renderedHtml);
            return ResponseEntity.ok(response);
        } catch (BlogNotFoundException ex) {
            return ResponseEntity.status(404).body(Map.of("error", ex.getMessage()));
        } catch (BlogAccessDeniedException ex) {
            return ResponseEntity.status(403).body(Map.of("error", ex.getMessage()));
        }
    }

    @PostMapping("/{id}/comments")
    public ResponseEntity<Map<String, Object>> addComment(
            @PathVariable String id,
            @RequestBody Map<String, String> body,
            @RequestHeader("X-Username") String username) {
        try {
            Blog blog = blogService.addComment(id, username, body.get("text"));
            return ResponseEntity.status(201).body(toBlogResponse(blog, username));
        } catch (BlogNotFoundException ex) {
            return ResponseEntity.status(404).body(Map.of("error", ex.getMessage()));
        } catch (BlogAccessDeniedException ex) {
            return ResponseEntity.status(403).body(Map.of("error", ex.getMessage()));
        }
    }

    @PutMapping("/{blogId}/comments/{commentId}")
    public ResponseEntity<?> editComment(
        @PathVariable String blogId,
        @PathVariable String commentId,
        @RequestBody Map<String, String> body,
        @RequestHeader("X-Username") String username){

            try{
                Blog blog = blogService.editComment(blogId, commentId, username, body.get("text"));
                return ResponseEntity.ok(toBlogResponse(blog, username));
            }catch(RuntimeException e){
                return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
            }
    }

    @PostMapping("/{id}/like")
    public ResponseEntity<Map<String, Object>> likeBlog(@PathVariable String id, @RequestHeader("X-Username") String username) {
        Blog blog = blogService.likeBlog(id, username);
        return ResponseEntity.ok(Map.of(
                "likesCount", blog.getLikes().size(),
                "likedByCurrentUser", blog.getLikes().contains(username)
        ));
    }

    private Map<String, Object> toBlogResponse(Blog blog, String username) {
        Map<String, Object> response = new HashMap<>();
        response.put("blog", toPublicBlog(blog));
        response.put("likesCount", blog.getLikes().size());
        response.put("likedByCurrentUser", username != null && blog.getLikes().contains(username));
        return response;
    }

    private Map<String, Object> toPublicBlog(Blog blog) {
        Map<String, Object> blogMap = new HashMap<>();
        blogMap.put("id", blog.getId());
        blogMap.put("title", blog.getTitle());
        blogMap.put("descriptionMarkdown", blog.getDescriptionMarkdown());
        blogMap.put("authorUsername", blog.getAuthorUsername());
        blogMap.put("createdAt", blog.getCreatedAt());
        blogMap.put("imageUrls", blog.getImageUrls());
        blogMap.put("comments", blog.getComments());
        return blogMap;
    }
}