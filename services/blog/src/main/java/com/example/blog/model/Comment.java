package com.example.blog.model;

import lombok.Data;
import java.time.LocalDateTime;
import java.util.UUID;

@Data
public class Comment{

    private String id = UUID.randomUUID().toString();

    private String authorUsername;

    private String text;

    private LocalDateTime createdAt = LocalDateTime.now();

    private LocalDateTime lastModifiedAt;
}