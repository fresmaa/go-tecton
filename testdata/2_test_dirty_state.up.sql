-- First query is valid and will be executed successfully
ALTER TABLE users ADD COLUMN email VARCHAR(100);

-- Second query is invalid and will cause an error
CREATE TABLE dummy_table (
    id INT PRIMARY KEY
    name VARCHAR(50) 
);