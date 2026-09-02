package com.example.blog.service;

import com.example.blog.client.FollowerClient;
import com.example.blog.client.NotificationClient;
import com.example.blog.exception.BlogAccessDeniedException;
import com.example.blog.exception.BlogNotFoundException;
import com.example.blog.model.Blog;
import com.example.blog.model.Comment;
import com.example.blog.repository.BlogRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;

import java.util.List;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.assertEquals;
import static org.junit.jupiter.api.Assertions.assertThrows;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.*;

// Plain Mockito unit tests, not @SpringBootTest - BlogService's
// author-only/comment-moderation checks are pure business logic and don't
// need a real Mongo connection or Spring context to exercise.
@ExtendWith(MockitoExtension.class)
class BlogServiceAuthorizationTest {

    @Mock
    private BlogRepository blogRepository;
    @Mock
    private FollowerClient followerClient;
    @Mock
    private NotificationClient notificationClient;

    @InjectMocks
    private BlogService blogService;

    private Blog blog;

    @BeforeEach
    void setUp() {
        blog = new Blog();
        blog.setId("blog-1");
        blog.setAuthorUsername("ana");
        blog.setTitle("Original title");
    }

    @Test
    void updateBlog_authorCanUpdateTheirOwnBlog() {
        when(blogRepository.findById("blog-1")).thenReturn(Optional.of(blog));
        when(blogRepository.save(any(Blog.class))).thenAnswer(inv -> inv.getArgument(0));

        Blog updated = blogService.updateBlog("blog-1", "ana", "New title", "New body", List.of());

        assertEquals("New title", updated.getTitle());
        verify(blogRepository).save(blog);
    }

    @Test
    void updateBlog_rejectsANonAuthor() {
        when(blogRepository.findById("blog-1")).thenReturn(Optional.of(blog));

        assertThrows(BlogAccessDeniedException.class, () ->
                blogService.updateBlog("blog-1", "marko", "New title", "New body", List.of()));
        verify(blogRepository, never()).save(any());
    }

    @Test
    void updateBlog_missingBlogIsNotFound() {
        when(blogRepository.findById("missing")).thenReturn(Optional.empty());

        assertThrows(BlogNotFoundException.class, () ->
                blogService.updateBlog("missing", "ana", "New title", "New body", List.of()));
    }

    @Test
    void deleteBlog_authorCanDeleteTheirOwnBlog() {
        when(blogRepository.findById("blog-1")).thenReturn(Optional.of(blog));

        blogService.deleteBlog("blog-1", "ana");

        verify(blogRepository).delete(blog);
    }

    @Test
    void deleteBlog_rejectsANonAuthor() {
        when(blogRepository.findById("blog-1")).thenReturn(Optional.of(blog));

        assertThrows(BlogAccessDeniedException.class, () -> blogService.deleteBlog("blog-1", "marko"));
        verify(blogRepository, never()).delete(any());
    }

    @Test
    void deleteComment_commentAuthorCanDeleteTheirOwnComment() {
        Comment comment = new Comment();
        comment.setId("c1");
        comment.setAuthorUsername("marko");
        blog.getComments().add(comment);

        when(blogRepository.findById("blog-1")).thenReturn(Optional.of(blog));
        when(blogRepository.save(any(Blog.class))).thenAnswer(inv -> inv.getArgument(0));

        Blog result = blogService.deleteComment("blog-1", "c1", "marko");

        assertEquals(0, result.getComments().size());
    }

    @Test
    void deleteComment_blogAuthorCanModerateSomeoneElsesComment() {
        Comment comment = new Comment();
        comment.setId("c1");
        comment.setAuthorUsername("marko");
        blog.getComments().add(comment);

        when(blogRepository.findById("blog-1")).thenReturn(Optional.of(blog));
        when(blogRepository.save(any(Blog.class))).thenAnswer(inv -> inv.getArgument(0));

        // blog.authorUsername is "ana" - not the comment's author, but the blog owner
        Blog result = blogService.deleteComment("blog-1", "c1", "ana");

        assertEquals(0, result.getComments().size());
    }

    @Test
    void deleteComment_rejectsAThirdPartyWhoIsNeitherAuthor() {
        Comment comment = new Comment();
        comment.setId("c1");
        comment.setAuthorUsername("marko");
        blog.getComments().add(comment);

        when(blogRepository.findById("blog-1")).thenReturn(Optional.of(blog));

        assertThrows(BlogAccessDeniedException.class, () ->
                blogService.deleteComment("blog-1", "c1", "stranger"));
        verify(blogRepository, never()).save(any());
    }
}
