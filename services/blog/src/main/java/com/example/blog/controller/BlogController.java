package com.example.blog.controller;

import com.example.blog.config.AuthUtil;
import com.example.blog.model.Blog;
import com.example.blog.service.BlogService;
import jakarta.servlet.http.HttpServletRequest;
import lombok.RequiredArgsConstructor;
import org.springframework.data.domain.Page;
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

    // Caller identity comes from the verified JWT (JwtAuthFilter), never a
    // client/gateway-supplied header. Exceptions are mapped centrally in
    // GlobalExceptionHandler.

    @PostMapping
    public ResponseEntity<Map<String, Object>> createBlog(@RequestBody Map<String, Object> body, HttpServletRequest request){
        String username = AuthUtil.requireUsername(request);

        String title = (String) body.get("title");
        String description = (String) body.get("descriptionMarkdown");
        List<String> images = (List<String>) body.get("imageUrls");

        if (title == null || title.isBlank() || description == null || description.isBlank()) {
            return ResponseEntity.badRequest().body(Map.of("error", "title and descriptionMarkdown are required"));
        }

        Blog blog = blogService.createBlog(title.trim(), description, images, username);
        return ResponseEntity.status(201).body(toBlogResponse(blog, username));
    }

    @GetMapping
    public ResponseEntity<Map<String, Object>> getAllBlogs(
            @RequestParam(defaultValue = "0") int page,
            @RequestParam(defaultValue = "10") int size,
            HttpServletRequest request){
        String username = AuthUtil.requireUsername(request);
        Page<Blog> result = blogService.getAllBlogsForUser(username, page, size);
        List<Map<String, Object>> blogs = result.getContent().stream()
                .map(blog -> toBlogResponse(blog, username))
                .collect(java.util.stream.Collectors.toList());
        return ResponseEntity.ok(Map.of(
                "blogs", blogs,
                "page", result.getNumber(),
                "size", result.getSize(),
                "totalElements", result.getTotalElements(),
                "totalPages", result.getTotalPages()
        ));
    }

    //dobavljanje jednog bloga (sa rendered markdown)
    @GetMapping("/{id}")
    public ResponseEntity<?> getBlogById(@PathVariable String id, HttpServletRequest request){
        String username = AuthUtil.requireUsername(request);
        Blog blog = blogService.getBlogByIdForUser(id, username);
        String renderedHtml = blogService.renderMarkdown(blog.getDescriptionMarkdown());
        Map<String, Object> response = toBlogResponse(blog, username);
        response.put("descriptionHtml", renderedHtml);
        return ResponseEntity.ok(response);
    }

    @PostMapping("/{id}/comments")
    public ResponseEntity<Map<String, Object>> addComment(
            @PathVariable String id,
            @RequestBody Map<String, String> body,
            HttpServletRequest request) {
        String username = AuthUtil.requireUsername(request);
        String text = body.get("text");
        if (text == null || text.isBlank()) {
            return ResponseEntity.badRequest().body(Map.of("error", "text is required"));
        }
        Blog blog = blogService.addComment(id, username, text.trim());
        return ResponseEntity.status(201).body(toBlogResponse(blog, username));
    }

    @PutMapping("/{blogId}/comments/{commentId}")
    public ResponseEntity<?> editComment(
        @PathVariable String blogId,
        @PathVariable String commentId,
        @RequestBody Map<String, String> body,
        HttpServletRequest request){
        String username = AuthUtil.requireUsername(request);
        String text = body.get("text");
        if (text == null || text.isBlank()) {
            return ResponseEntity.badRequest().body(Map.of("error", "text is required"));
        }
        Blog blog = blogService.editComment(blogId, commentId, username, text.trim());
        return ResponseEntity.ok(toBlogResponse(blog, username));
    }

    @PutMapping("/{id}")
    public ResponseEntity<Map<String, Object>> updateBlog(
            @PathVariable String id,
            @RequestBody Map<String, Object> body,
            HttpServletRequest request){
        String username = AuthUtil.requireUsername(request);
        String title = (String) body.get("title");
        String description = (String) body.get("descriptionMarkdown");
        List<String> images = (List<String>) body.get("imageUrls");

        if (title == null || title.isBlank() || description == null || description.isBlank()) {
            return ResponseEntity.badRequest().body(Map.of("error", "title and descriptionMarkdown are required"));
        }

        Blog blog = blogService.updateBlog(id, username, title.trim(), description, images);
        return ResponseEntity.ok(toBlogResponse(blog, username));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Map<String, String>> deleteBlog(@PathVariable String id, HttpServletRequest request){
        String username = AuthUtil.requireUsername(request);
        blogService.deleteBlog(id, username);
        return ResponseEntity.ok(Map.of("message", "Blog obrisan"));
    }

    @DeleteMapping("/{blogId}/comments/{commentId}")
    public ResponseEntity<Map<String, Object>> deleteComment(
            @PathVariable String blogId,
            @PathVariable String commentId,
            HttpServletRequest request){
        String username = AuthUtil.requireUsername(request);
        Blog blog = blogService.deleteComment(blogId, commentId, username);
        return ResponseEntity.ok(toBlogResponse(blog, username));
    }

    @PostMapping("/{id}/like")
    public ResponseEntity<Map<String, Object>> likeBlog(@PathVariable String id, HttpServletRequest request) {
        String username = AuthUtil.requireUsername(request);
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
