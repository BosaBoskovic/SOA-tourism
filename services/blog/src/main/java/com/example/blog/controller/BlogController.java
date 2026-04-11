package com.example.blog.controller;

import com.example.blog.model.Blog;
import com.example.blog.service.BlogService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

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
    public ResponseEntity<Blog> createBlog(@RequestBody Map<String, Object> body, @RequestHeader("X-Username") String username){
        
        String title = (String) body.get("title");
        String description = (String) body.get("descriptionMarkdown");
        List<String> images = (List<String>) body.get("imageUrls");

        Blog blog = blogService.createBlog(title, description, images, username);
        return ResponseEntity.status(201).body(blog);
    }

    //dobavljanje svih blogova
    @GetMapping
    public ResponseEntity<List<Blog>> getAllBlogs(){
        return ResponseEntity.ok(blogService.getAllBlogs());
    }

    //dobavljanje jednog bloga (sa rendered markdown)
    @GetMapping("/{id}")
    public ResponseEntity<?>getBlogById(@PathVariable String id){
        return blogService.getBlogById(id).map(blog -> {
            String renderedHtml = blogService.renderMarkdown(blog.getDescriptionMarkdown());
            return ResponseEntity.ok(Map.of(
                "blog", blog,
                "descriptionHtml", renderedHtml
            ));
        })
        .orElse(ResponseEntity.notFound().build());
    }

    @PostMapping("/{id}/comments")
    public ResponseEntity<Blog> addComment(
            @PathVariable String id,
            @RequestBody Map<String, String> body,
            @RequestHeader("X-Username") String username) {
        Blog blog = blogService.addComment(id, username, body.get("text"));
        return ResponseEntity.status(201).body(blog);
    }

    @PutMapping("/{blogId}/comments/{commentId}")
    public ResponseEntity<?> editComment(
        @PathVariable String blogId,
        @PathVariable String commentId,
        @RequestBody Map<String, String> body,
        @RequestHeader("X-Username") String username){

            try{
                Blog blog = blogService.editComment(blogId, commentId, username, body.get("text"));
                return ResponseEntity.ok(blog);
            }catch(RuntimeException e){
                return ResponseEntity.badRequest().body(Map.of("error", e.getMessage()));
            }
    }
    
}