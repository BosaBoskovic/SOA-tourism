package com.example.blog.repository;

import com.example.blog.model.Blog;
import org.springframework.data.mongodb.repository.MongoRepository;

public interface BlogRepository extends MongoRepository<Blog, String>{
    
}