package com.example.blog.exception;

public class BlogAccessDeniedException extends RuntimeException {
    public BlogAccessDeniedException(String message) {
        super(message);
    }
}
