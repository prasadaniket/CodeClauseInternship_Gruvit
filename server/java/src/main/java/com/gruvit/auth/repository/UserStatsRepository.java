package com.gruvit.auth.repository;

import com.gruvit.auth.entity.UserStats;
import org.springframework.data.mongodb.repository.MongoRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface UserStatsRepository extends MongoRepository<UserStats, String> {
    Optional<UserStats> findByUserId(String userId);
    void deleteByUserId(String userId);
}
