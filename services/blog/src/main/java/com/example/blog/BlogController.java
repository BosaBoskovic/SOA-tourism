package com.example.blog;

import org.springframework.web.bind.annotation.*;
import java.util.Map;

@RestController
public class BlogController {

    @GetMapping("/")
    public String home() {
        return "Blog API radi";
    }

    @GetMapping("/blog")
    public Map<String, String> blog() {
        return Map.of("message", "Ovo je blog endpoint");
    }
}