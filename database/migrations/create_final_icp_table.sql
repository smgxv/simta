CREATE TABLE `final_icp` (
  `id` int NOT NULL AUTO_INCREMENT,
  `user_id` int NOT NULL,
  `dosen_id` int NOT NULL,
  `nama_lengkap` varchar(255) NOT NULL,
  `jurusan` varchar(255) NOT NULL,
  `kelas` varchar(50) NOT NULL,
  `topik_penelitian` varchar(255) NOT NULL,
  `file_path` varchar(255) NOT NULL,
  `keterangan` text,
  `status` enum('pending','approved','rejected','on review') NOT NULL DEFAULT 'pending',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  KEY `user_id` (`user_id`),
  KEY `dosen_id` (`dosen_id`),
  CONSTRAINT `final_icp_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,
  CONSTRAINT `final_icp_ibfk_2` FOREIGN KEY (`dosen_id`) REFERENCES `dosen` (`id`) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci; 