SELECT COUNT(*)
FROM
	(SELECT DISTINCT U."authorId"
		FROM URLS U)
