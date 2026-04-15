# BookMyShow - DB Design

## Problem
Design a database for a BookMyShow-like platform. No login required — users browse as guests. They select a city, see the latest movies running, and on clicking a movie they see the list of theatres showing it. Filters available: language and format (2D, 3D, IMAX).

## Requirements
1. Get all cities
2. Get latest movies running in a selected city (movies that have shows scheduled this week)
3. Get list of theatres where a movie is running in the current week
4. Filter by language and format

## Schema

```
cities:        id, name, short_name, image_url
movies:        id, name, image_url, director, cast
theatres:      id, name, city_id, is_active
languages:     id, name
format:        id, name
genre:         id, genre_name
movie_genres:  id, movie_id, genre_id
shows:         id, theatre_id, movie_id, format_id, language_id, date, time
```

- `shows` is the central table — language and format live here because they vary per screening, not per movie.
- No `is_latest` flag on movies. A movie is "latest" if it has shows this week. Derive from data, don't store it.
- City is derived from `theatres.city_id`, so no `city_id` on shows.

## Queries

### 1. Get all cities
```sql
SELECT * FROM cities;
```

### 2. Get latest movies running in a city this week
```sql
SELECT DISTINCT m.*
FROM movies m
JOIN shows s ON s.movie_id = m.id
JOIN theatres t ON t.id = s.theatre_id
WHERE t.city_id = ?
  AND s.date BETWEEN CURDATE() AND DATE_ADD(CURDATE(), INTERVAL 7 DAY);
```

### 3. Theatres showing a movie with language + format filters
```sql
SELECT DISTINCT t.*, s.date, s.time, l.name AS language, f.name AS format
FROM theatres t
JOIN shows s ON s.theatre_id = t.id
JOIN languages l ON l.id = s.language_id
JOIN format f ON f.id = s.format_id
WHERE t.city_id = ?
  AND s.movie_id = ?
  AND s.date BETWEEN CURDATE() AND DATE_ADD(CURDATE(), INTERVAL 7 DAY)
  AND s.language_id = ?    -- optional filter
  AND s.format_id = ?;     -- optional filter
```

## Indexing Notes
`shows` will be the most queried table. Composite indexes on `(theatre_id, movie_id, date)` and `(movie_id, date)` will help at scale.
