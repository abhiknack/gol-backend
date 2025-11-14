-- Initialize PostgreSQL Database for Supabase Redis Middleware
-- This script creates sample tables and data for development/testing

-- Supermarket Products Table
CREATE TABLE IF NOT EXISTS supermarket_products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10, 2),
    stock INTEGER DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_supermarket_products_category ON supermarket_products(category);

-- Movies Table
CREATE TABLE IF NOT EXISTS movies (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    genre VARCHAR(100),
    duration INTEGER,
    rating DECIMAL(3, 1),
    release_date DATE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_movies_genre ON movies(genre);

-- Movie Showtimes Table
CREATE TABLE IF NOT EXISTS showtimes (
    id SERIAL PRIMARY KEY,
    movie_id INTEGER REFERENCES movies(id) ON DELETE CASCADE,
    theater VARCHAR(255),
    showtime TIMESTAMP,
    available_seats INTEGER,
    price DECIMAL(10, 2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_showtimes_movie_id ON showtimes(movie_id);
CREATE INDEX IF NOT EXISTS idx_showtimes_showtime ON showtimes(showtime);

-- Pharmacy Medicines Table
CREATE TABLE IF NOT EXISTS medicines (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    category VARCHAR(100),
    price DECIMAL(10, 2),
    prescription_required BOOLEAN DEFAULT false,
    stock INTEGER DEFAULT 0,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_medicines_category ON medicines(category);

-- Insert Sample Data for Supermarket Products
INSERT INTO supermarket_products (name, category, price, stock, description) VALUES
    ('Whole Milk', 'dairy', 3.99, 50, 'Fresh whole milk, 1 gallon'),
    ('White Bread', 'bakery', 2.49, 100, 'Freshly baked white bread'),
    ('Brown Eggs', 'dairy', 4.99, 75, 'Farm fresh brown eggs, dozen'),
    ('Cheddar Cheese', 'dairy', 5.99, 40, 'Sharp cheddar cheese, 8oz'),
    ('Organic Apples', 'produce', 4.49, 120, 'Organic red apples, per lb'),
    ('Bananas', 'produce', 1.99, 200, 'Fresh bananas, per lb'),
    ('Ground Beef', 'meat', 8.99, 30, 'Fresh ground beef, per lb'),
    ('Chicken Breast', 'meat', 7.99, 45, 'Boneless chicken breast, per lb'),
    ('Orange Juice', 'beverages', 4.99, 60, 'Fresh squeezed orange juice, 64oz'),
    ('Coffee Beans', 'beverages', 12.99, 35, 'Premium coffee beans, 12oz')
ON CONFLICT DO NOTHING;

-- Insert Sample Data for Movies
INSERT INTO movies (title, genre, duration, rating, release_date, description) VALUES
    ('Action Hero Returns', 'action', 120, 8.5, '2024-01-15', 'The hero returns for one last mission'),
    ('Comedy Night Live', 'comedy', 95, 7.8, '2024-02-01', 'A hilarious stand-up comedy special'),
    ('Mystery Manor', 'thriller', 110, 8.2, '2024-01-20', 'A mysterious murder in an old manor'),
    ('Space Odyssey 2024', 'sci-fi', 140, 9.0, '2024-03-01', 'An epic journey through space'),
    ('Love in Paris', 'romance', 105, 7.5, '2024-02-14', 'A romantic tale set in Paris'),
    ('Horror House', 'horror', 90, 7.0, '2024-10-31', 'A terrifying haunted house experience'),
    ('Family Adventure', 'family', 100, 8.0, '2024-04-01', 'A fun adventure for the whole family'),
    ('Documentary: Earth', 'documentary', 85, 8.8, '2024-05-01', 'Exploring the wonders of our planet')
ON CONFLICT DO NOTHING;

-- Insert Sample Data for Showtimes
INSERT INTO showtimes (movie_id, theater, showtime, available_seats, price) VALUES
    (1, 'Cinema Plaza', '2024-11-15 14:00:00', 150, 12.99),
    (1, 'Cinema Plaza', '2024-11-15 17:30:00', 120, 12.99),
    (1, 'Cinema Plaza', '2024-11-15 20:00:00', 100, 14.99),
    (2, 'Megaplex Theater', '2024-11-15 15:00:00', 200, 10.99),
    (2, 'Megaplex Theater', '2024-11-15 19:00:00', 180, 10.99),
    (3, 'Cinema Plaza', '2024-11-15 16:00:00', 130, 12.99),
    (3, 'Downtown Cinema', '2024-11-15 18:30:00', 90, 11.99),
    (4, 'IMAX Theater', '2024-11-15 13:00:00', 250, 18.99),
    (4, 'IMAX Theater', '2024-11-15 19:30:00', 220, 18.99)
ON CONFLICT DO NOTHING;

-- Insert Sample Data for Medicines
INSERT INTO medicines (name, category, price, prescription_required, stock, description) VALUES
    ('Aspirin 500mg', 'pain-relief', 5.99, false, 200, 'Pain relief and fever reducer'),
    ('Ibuprofen 200mg', 'pain-relief', 7.99, false, 180, 'Anti-inflammatory pain reliever'),
    ('Amoxicillin 500mg', 'antibiotic', 12.99, true, 50, 'Antibiotic for bacterial infections'),
    ('Vitamin C 1000mg', 'vitamins', 9.99, false, 150, 'Immune system support'),
    ('Multivitamin', 'vitamins', 14.99, false, 120, 'Daily multivitamin supplement'),
    ('Cough Syrup', 'cold-flu', 8.99, false, 100, 'Relief for cough and cold symptoms'),
    ('Allergy Relief', 'allergy', 11.99, false, 90, 'Non-drowsy allergy relief'),
    ('Blood Pressure Monitor', 'medical-devices', 49.99, false, 25, 'Digital blood pressure monitor'),
    ('First Aid Kit', 'first-aid', 24.99, false, 40, 'Complete first aid kit'),
    ('Thermometer', 'medical-devices', 12.99, false, 60, 'Digital thermometer')
ON CONFLICT DO NOTHING;

-- Create a view for available movies with showtimes
CREATE OR REPLACE VIEW available_movies AS
SELECT 
    m.id,
    m.title,
    m.genre,
    m.duration,
    m.rating,
    COUNT(s.id) as showtime_count,
    MIN(s.price) as min_price,
    MAX(s.price) as max_price
FROM movies m
LEFT JOIN showtimes s ON m.id = s.movie_id
GROUP BY m.id, m.title, m.genre, m.duration, m.rating;

-- Create a view for low stock products
CREATE OR REPLACE VIEW low_stock_products AS
SELECT 
    'supermarket' as domain,
    id,
    name,
    category,
    stock,
    price
FROM supermarket_products
WHERE stock < 50
UNION ALL
SELECT 
    'pharmacy' as domain,
    id,
    name,
    category,
    stock,
    price
FROM medicines
WHERE stock < 100;

-- Grant permissions (if needed)
-- GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
-- GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO postgres;

-- Display summary
SELECT 'Database initialized successfully!' as status;
SELECT 'Supermarket Products: ' || COUNT(*) as count FROM supermarket_products;
SELECT 'Movies: ' || COUNT(*) as count FROM movies;
SELECT 'Showtimes: ' || COUNT(*) as count FROM showtimes;
SELECT 'Medicines: ' || COUNT(*) as count FROM medicines;
