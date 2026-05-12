package com.example.blog.client;

import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Component;
import org.springframework.web.client.RestClient;

import java.util.HashSet;
import java.util.List;
import java.util.Map;
import java.util.Set;

@Component
public class FollowerClient {

    private final RestClient restClient;

    public FollowerClient(@Value("${follower.service.url}") String followerServiceUrl) {
        this.restClient = RestClient.builder().baseUrl(followerServiceUrl).build();
    }

    public boolean isFollowing(String followerUsername, String targetUsername) {
        if (followerUsername == null || targetUsername == null) {
            return false;
        }

        if (followerUsername.equals(targetUsername)) {
            return true;
        }

        Map<String, Object> response = restClient.get()
                .uri(uriBuilder -> uriBuilder
                        .path("/followers/is-following")
                        .queryParam("followerUsername", followerUsername)
                        .queryParam("targetUsername", targetUsername)
                        .build())
                .retrieve()
                .body(Map.class);

        if (response == null || !response.containsKey("isFollowing")) {
            return false;
        }

        Object value = response.get("isFollowing");
        if (value instanceof Boolean follows) {
            return follows;
        }

        return Boolean.parseBoolean(String.valueOf(value));
    }

    public Set<String> getVisibleAuthors(String username) {
        if (username == null || username.isBlank()) {
            return Set.of();
        }

        Map<String, Object> response = restClient.get()
                .uri("/followers/visible-authors/{username}", username)
                .retrieve()
                .body(Map.class);

        if (response == null) {
            return Set.of(username);
        }

        Object rawAuthors = response.get("authors");
        if (!(rawAuthors instanceof List<?> authorsList)) {
            return Set.of(username);
        }

        Set<String> authors = new HashSet<>();
        for (Object item : authorsList) {
            if (item != null) {
                authors.add(String.valueOf(item));
            }
        }

        authors.add(username);
        return authors;
    }
}
