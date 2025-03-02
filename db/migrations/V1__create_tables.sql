-- Create the restaurants table.
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'restaurants')
BEGIN
  CREATE TABLE restaurants (
    id INT IDENTITY(1,1) PRIMARY KEY,
    name NVARCHAR(255) NOT NULL,
    style NVARCHAR(255) NOT NULL,
    address NVARCHAR(255),
    openHour NVARCHAR(5) NOT NULL,
    closeHour NVARCHAR(5) NOT NULL,
    vegetarian BIT,
    deliveries BIT
  );
END;

-- Create the query_logs table.
IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = 'query_logs')
BEGIN
  CREATE TABLE query_logs (
    id INT IDENTITY(1,1) PRIMARY KEY,
    query NVARCHAR(MAX) NOT NULL,
    response NVARCHAR(MAX) NOT NULL,
    created_at DATETIME DEFAULT GETDATE()
  );
END;