package com.example.blog.rpc;

import com.example.blog.exception.BlogAccessDeniedException;
import com.example.blog.exception.BlogNotFoundException;
import com.example.blog.model.Blog;
import com.example.blog.service.BlogService;
import lombok.RequiredArgsConstructor;
import blogs.v1.BlogsServiceGrpc;
import blogs.v1.GetAllBlogsRequest;
import blogs.v1.GetAllBlogsResponse;
import blogs.v1.GetBlogRequest;
import blogs.v1.GetBlogResponse;
import io.grpc.Status;
import io.grpc.stub.StreamObserver;
import net.devh.boot.grpc.server.service.GrpcService;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.util.List;

@GrpcService
@RequiredArgsConstructor
public class BlogGrpcService extends BlogsServiceGrpc.BlogsServiceImplBase {

    private static final Logger log = LoggerFactory.getLogger(BlogGrpcService.class);

    private final BlogService blogService;

    @Override
    public void getBlog(GetBlogRequest request,
                        StreamObserver<GetBlogResponse> responseObserver) {
        log.info("RPC GetBlog called, blogId={}", request.getBlogId());

        try {
            Blog blog = blogService.getBlogByIdForUser(
                    request.getBlogId(),
                    request.getUsername()
            );

            GetBlogResponse response = GetBlogResponse.newBuilder()
                    .setBlog(toGrpcBlog(blog, request.getUsername(), true))
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (BlogNotFoundException ex) {
            responseObserver.onError(Status.NOT_FOUND.withDescription(ex.getMessage()).asRuntimeException());
        } catch (BlogAccessDeniedException ex) {
            responseObserver.onError(Status.PERMISSION_DENIED.withDescription(ex.getMessage()).asRuntimeException());
        } catch (Exception ex) {
            log.error("GetBlog failed", ex);
            responseObserver.onError(Status.INTERNAL.withDescription("internal error").asRuntimeException());
        }
    }

    @Override
    public void getAllBlogs(GetAllBlogsRequest request,
                            StreamObserver<GetAllBlogsResponse> responseObserver) {
        log.info("RPC GetAllBlogs called, username={}", request.getUsername());

        try {
            List<Blog> blogs = blogService.getAllBlogsForUser(request.getUsername());

            GetAllBlogsResponse response = GetAllBlogsResponse.newBuilder()
                    .addAllBlogs(
                            blogs.stream()
                                    .map(blog -> toGrpcBlog(blog, request.getUsername(), false))
                                    .toList()
                    )
                    .build();

            responseObserver.onNext(response);
            responseObserver.onCompleted();
        } catch (Exception ex) {
            log.error("GetAllBlogs failed", ex);
            responseObserver.onError(Status.INTERNAL.withDescription("internal error").asRuntimeException());
        }
    }

    private blogs.v1.Blog toGrpcBlog(Blog blog, String username, boolean includeHtml) {
        blogs.v1.Blog.Builder builder = blogs.v1.Blog.newBuilder()
                .setId(blog.getId())
                .setTitle(blog.getTitle())
                .setDescriptionMarkdown(blog.getDescriptionMarkdown())
                .setAuthorUsername(blog.getAuthorUsername())
                .setCreatedAt(blog.getCreatedAt().toString())
                .addAllImageUrls(blog.getImageUrls())
                .setLikesCount(blog.getLikes().size())
                .setLikedByCurrentUser(blog.getLikes().contains(username));

        if (includeHtml) {
            builder.setDescriptionHtml(
                    blogService.renderMarkdown(blog.getDescriptionMarkdown())
            );
        }

        return builder.build();
    }
}
