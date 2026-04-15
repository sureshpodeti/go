Bookmyshow

1. Get all cities - select *from cities;
2. Get latest movies running - select *from movies where is_latest=true;
3. Get list of theatres where the movie is running in any day in the current week - 
4. Filter by - language, format

city: theatre
1:M


cities
    id
    name
    short_name
    icon_url

movies
    id 
    name
    language_id
    genre_id
    is_latest

language
    id
    name

genres
    id 
    name


theatres
    id
    name
    city_id

shows
    id 
    city_id
    theatre_id
    movie_id
    date

movie_genre


movie_language


select *from theaters inner join on shows 

Queries:
    Get 


TODO:
Location sensitive show/movies display


Assume:
No audium type theatres considered