package com.example.blog.repository;

import com.example.blog.model.Blog;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.mongodb.repository.MongoRepository;

import java.util.Collection;

public interface BlogRepository extends MongoRepository<Blog, String>{

    // Pages at the DB level, not in memory - the visible-authors filter
    // happens in the WHERE clause (author IN (...)) instead of fetch-everything-then-filter.
    Page<Blog> findByAuthorUsernameIn(Collection<String> authorUsernames, Pageable pageable);
}