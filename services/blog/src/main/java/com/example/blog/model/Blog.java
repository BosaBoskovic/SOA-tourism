package com.example.blog.model;

import lombok.Data;
import org.springframework.data.annotation.Id;
import org.springframework.data.mongodb.core.mapping.Document;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashSet;
import java.util.List;
import java.util.Set;

@Data
@Document(collection = "blogs")
public class Blog{

    @Id
    private String id;

    private String title;

    //cuvam markdown, renderujem po potrebi
    private String descriptionMarkdown;

    private String authorUsername;

    private LocalDateTime createdAt = LocalDateTime.now();

    //opcione slike
    private List<String> imageUrls = new ArrayList<>();

    private List<Comment> comments = new ArrayList<>();

    private Set<String> likes = new HashSet<>();
}