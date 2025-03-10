package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/soa-rs/fit/internal/config"
	"github.com/soa-rs/fit/internal/config/logger"
)

// Models based on DB schema
type User struct {
	ID          int       `json:"id"`
	Email       string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Exercise struct {
	ID              int      `json:"id"`
	Name            string   `json:"name"`
	Equipment       []string `json:"equipment"`
	PrimaryMuscles  []string `json:"primary_muscles"`
	SecondaryMuscles []string `json:"secondary_muscles"`
	ExerciseType    string   `json:"exercise_type"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type Program struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Name      string    `json:"name"`
	IsPublic  bool      `json:"is_public"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Routine struct {
	ID        int       `json:"id"`
	ProgramID int       `json:"program_id"`
	Name      string    `json:"name"`
	DayNumber int       `json:"day_number"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RoutineExercise struct {
	ID                 int     `json:"id"`
	RoutineID          int     `json:"routine_id"`
	ExerciseID         int     `json:"exercise_id"`
	RecommendedSets    int     `json:"recommended_sets"`
	RecommendedReps    int     `json:"recommended_reps"`
	RecommendedRPE     float64 `json:"recommended_rpe"`
	RecommendedDuration int     `json:"recommended_duration"`
	RecommendedDistance float64 `json:"recommended_distance"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type Workout struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	RoutineID   int       `json:"routine_id"`
	PerformedAt time.Time `json:"performed_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkoutSet struct {
	ID         int     `json:"id"`
	WorkoutID  int     `json:"workout_id"`
	ExerciseID int     `json:"exercise_id"`
	Sets       int     `json:"sets"`
	Reps       int     `json:"reps"`
	Weight     float64 `json:"weight"`
	RPE        float64 `json:"rpe"`
	Duration   int     `json:"duration"`
	Distance   float64 `json:"distance"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// Database connection
var db *sql.DB

// Init database connection
func initDB() {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.GetEnvOrDefault("DB_HOST"),
		config.GetEnvOrDefault("DB_PORT"),
		config.GetEnvOrDefault("DB_USER"),
		config.GetEnvOrDefault("DB_PASSWORD"),
		config.GetEnvOrDefault("DB_NAME"),
	)

	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		logger.LogFatal("Failed to connect to database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		logger.LogFatal("Failed to ping database: %v", err)
	}

	logger.LogInfo("Connected to database successfully")
}

// Helper functions for pagination
func getPaginationParams(c *gin.Context) (int, int) {
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}
	
	offset := (page - 1) * limit
	return limit, offset
}

// Main function
func main() {
	config.LoadEnvs()
	config.SetupLogger()
	
	// Initialize database
	initDB()
	
	// Initialize Gin
	router := gin.Default()
	router.SetTrustedProxies(nil)
	
	// Health check route
	router.GET("/health", func(c *gin.Context) {
		logger.LogTrace("Health check route")
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	
	// API Routes
	api := router.Group("/api")
	{
		// Exercise routes (Milestone 2)
		exercises := api.Group("/exercises")
		{
			exercises.POST("", createExercise)
			exercises.GET("", listExercises)
			exercises.GET("/:id", getExerciseByID)
			exercises.PUT("/:id", updateExercise)
			exercises.DELETE("/:id", deleteExercise)
		}
		
		// Program routes (Milestone 3)
		programs := api.Group("/programs")
		{
			programs.POST("", createProgram)
			programs.GET("", listPrograms)
			programs.GET("/:id", getProgramByID)
			programs.PUT("/:id", updateProgram)
			programs.DELETE("/:id", deleteProgram)
		}
		
		// Routine routes (Milestone 3)
		routines := api.Group("/routines")
		{
			routines.POST("", createRoutine)
			routines.GET("", listRoutines)
			routines.GET("/:id", getRoutineByID)
			routines.PUT("/:id", updateRoutine)
			routines.DELETE("/:id", deleteRoutine)
			
			// Routine-Exercise linking
			routines.POST("/:id/exercises", addExerciseToRoutine)
			routines.DELETE("/:id/exercises/:exerciseId", removeExerciseFromRoutine)
			routines.GET("/:id/exercises", getRoutineExercises)
			routines.PUT("/:id/exercises/:exerciseId", updateRoutineExercise)
		}
		
		// Workout routes (Milestone 4)
		workouts := api.Group("/workouts")
		{
			workouts.POST("", createWorkout)
			workouts.GET("", listWorkouts)
			workouts.GET("/:id", getWorkoutByID)
			
			// Workout sets
			workouts.POST("/:id/sets", addWorkoutSet)
			workouts.GET("/:id/sets", getWorkoutSets)
			workouts.PUT("/:id/sets/:setId", updateWorkoutSet)
			// TODO
			workouts.DELETE("/:id/sets/:setId", deleteWorkoutSet)
		}
	}
	
	// Start server
	port := config.GetEnvOrDefault(config.EnvBackendPort)
	host := config.GetEnvOrDefault(config.EnvBackendHost)
	if err := router.Run(fmt.Sprintf("%s:%s", host, port)); err != nil {
		logger.LogFatal("Failed to run server: %v", err)
	}
}

// -------------------- Exercise Handlers (Milestone 2) --------------------

func createExercise(c *gin.Context) {
	var exercise Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validation
	if exercise.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	
	if exercise.ExerciseType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise type is required"})
		return
	}
	
	// Insert into database
	query := `
		INSERT INTO exercises (name, equipment, primary_muscles, secondary_muscles, exercise_type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	
	err := db.QueryRow(
		query,
		exercise.Name,
		exercise.Equipment,
		exercise.PrimaryMuscles,
		exercise.SecondaryMuscles,
		exercise.ExerciseType,
	).Scan(&exercise.ID, &exercise.CreatedAt, &exercise.UpdatedAt)
	
	if err != nil {
		logger.LogError("Failed to create exercise: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create exercise"})
		return
	}
	
	c.JSON(http.StatusCreated, exercise)
}

func getExerciseByID(c *gin.Context) {
	id := c.Param("id")
	
	var exercise Exercise
	query := `
		SELECT id, name, equipment, primary_muscles, secondary_muscles, exercise_type, created_at, updated_at
		FROM exercises
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Equipment,
		&exercise.PrimaryMuscles,
		&exercise.SecondaryMuscles,
		&exercise.ExerciseType,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
			return
		}
		
		logger.LogError("Failed to get exercise: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get exercise"})
		return
	}
	
	c.JSON(http.StatusOK, exercise)
}

func listExercises(c *gin.Context) {
	limit, offset := getPaginationParams(c)
	
	// Optional filtering by type
	exerciseType := c.Query("type")
	var whereClause string
	var args []interface{}
	
	if exerciseType != "" {
		whereClause = "WHERE exercise_type = $3"
		args = append(args, exerciseType)
	}
	
	query := fmt.Sprintf(`
		SELECT id, name, equipment, primary_muscles, secondary_muscles, exercise_type, created_at, updated_at
		FROM exercises
		%s
		ORDER BY name
		LIMIT $1 OFFSET $2
	`, whereClause)
	
	args = append([]interface{}{limit, offset}, args...)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		logger.LogError("Failed to list exercises: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list exercises"})
		return
	}
	defer rows.Close()
	
	exercises := []Exercise{}
	for rows.Next() {
		var exercise Exercise
		if err := rows.Scan(
			&exercise.ID,
			&exercise.Name,
			&exercise.Equipment,
			&exercise.PrimaryMuscles,
			&exercise.SecondaryMuscles,
			&exercise.ExerciseType,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
		); err != nil {
			logger.LogError("Failed to scan exercise row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process exercises"})
			return
		}
		exercises = append(exercises, exercise)
	}
	
	// Get total count for pagination info
	var total int
	countQuery := "SELECT COUNT(*) FROM exercises"
	if exerciseType != "" {
		countQuery += " WHERE exercise_type = $1"
		err = db.QueryRow(countQuery, exerciseType).Scan(&total)
	} else {
		err = db.QueryRow(countQuery).Scan(&total)
	}
	
	if err != nil {
		logger.LogError("Failed to count exercises: %v", err)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": exercises,
		"pagination": gin.H{
			"total": total,
			"limit": limit,
			"offset": offset,
		},
	})
}

func updateExercise(c *gin.Context) {
	id := c.Param("id")
	
	// Check if exercise exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if exercise exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}
	
	// Parse request body
	var exercise Exercise
	if err := c.ShouldBindJSON(&exercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validation
	if exercise.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	
	if exercise.ExerciseType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise type is required"})
		return
	}
	
	// Update in database
	query := `
		UPDATE exercises
		SET name = $1, equipment = $2, primary_muscles = $3, secondary_muscles = $4, exercise_type = $5, updated_at = NOW()
		WHERE id = $6
		RETURNING id, name, equipment, primary_muscles, secondary_muscles, exercise_type, created_at, updated_at
	`
	
	err = db.QueryRow(
		query,
		exercise.Name,
		exercise.Equipment,
		exercise.PrimaryMuscles,
		exercise.SecondaryMuscles,
		exercise.ExerciseType,
		id,
	).Scan(
		&exercise.ID,
		&exercise.Name,
		&exercise.Equipment,
		&exercise.PrimaryMuscles,
		&exercise.SecondaryMuscles,
		&exercise.ExerciseType,
		&exercise.CreatedAt,
		&exercise.UpdatedAt,
	)
	
	if err != nil {
		logger.LogError("Failed to update exercise: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update exercise"})
		return
	}
	
	c.JSON(http.StatusOK, exercise)
}

func deleteExercise(c *gin.Context) {
	id := c.Param("id")
	
	// Check if exercise exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if exercise exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found"})
		return
	}
	
	// Delete from database
	_, err = db.Exec("DELETE FROM exercises WHERE id = $1", id)
	if err != nil {
		logger.LogError("Failed to delete exercise: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete exercise"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Exercise deleted successfully"})
}

// -------------------- Program Handlers (Milestone 3) --------------------

func createProgram(c *gin.Context) {
	var program Program
	if err := c.ShouldBindJSON(&program); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validation
	if program.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	
	// Insert into database
	query := `
		INSERT INTO programs (user_id, name, is_public)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	
	err := db.QueryRow(
		query,
		program.UserID,
		program.Name,
		program.IsPublic,
	).Scan(&program.ID, &program.CreatedAt, &program.UpdatedAt)
	
	if err != nil {
		logger.LogError("Failed to create program: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create program"})
		return
	}
	
	c.JSON(http.StatusCreated, program)
}

func getProgramByID(c *gin.Context) {
	id := c.Param("id")
	
	var program Program
	query := `
		SELECT id, user_id, name, is_public, created_at, updated_at
		FROM programs
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&program.ID,
		&program.UserID,
		&program.Name,
		&program.IsPublic,
		&program.CreatedAt,
		&program.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
			return
		}
		
		logger.LogError("Failed to get program: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get program"})
		return
	}
	
	c.JSON(http.StatusOK, program)
}

func listPrograms(c *gin.Context) {
	limit, offset := getPaginationParams(c)
	userID := c.Query("user_id")
	
	var whereClause string
	var args []interface{}
	
	if userID != "" {
		whereClause = "WHERE user_id = $3 OR is_public = true"
		args = append(args, userID)
	} else {
		whereClause = "WHERE is_public = true"
	}
	
	query := fmt.Sprintf(`
		SELECT id, user_id, name, is_public, created_at, updated_at
		FROM programs
		%s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, whereClause)
	
	args = append([]interface{}{limit, offset}, args...)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		logger.LogError("Failed to list programs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list programs"})
		return
	}
	defer rows.Close()
	
	programs := []Program{}
	for rows.Next() {
		var program Program
		if err := rows.Scan(
			&program.ID,
			&program.UserID,
			&program.Name,
			&program.IsPublic,
			&program.CreatedAt,
			&program.UpdatedAt,
		); err != nil {
			logger.LogError("Failed to scan program row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process programs"})
			return
		}
		programs = append(programs, program)
	}
	
	// Get total count for pagination info
	var total int
	countQuery := "SELECT COUNT(*) FROM programs "
	if userID != "" {
		countQuery += "WHERE user_id = $1 OR is_public = true"
		err = db.QueryRow(countQuery, userID).Scan(&total)
	} else {
		countQuery += "WHERE is_public = true"
		err = db.QueryRow(countQuery).Scan(&total)
	}
	
	if err != nil {
		logger.LogError("Failed to count programs: %v", err)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": programs,
		"pagination": gin.H{
			"total": total,
			"limit": limit,
			"offset": offset,
		},
	})
}

func updateProgram(c *gin.Context) {
	id := c.Param("id")
	
	// Check if program exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM programs WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if program exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
		return
	}
	
	// Parse request body
	var program Program
	if err := c.ShouldBindJSON(&program); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validation
	if program.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	
	// Update in database
	query := `
		UPDATE programs
		SET name = $1, is_public = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, user_id, name, is_public, created_at, updated_at
	`
	
	err = db.QueryRow(
		query,
		program.Name,
		program.IsPublic,
		id,
	).Scan(
		&program.ID,
		&program.UserID,
		&program.Name,
		&program.IsPublic,
		&program.CreatedAt,
		&program.UpdatedAt,
	)
	
	if err != nil {
		logger.LogError("Failed to update program: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update program"})
		return
	}
	
	c.JSON(http.StatusOK, program)
}

func deleteProgram(c *gin.Context) {
	id := c.Param("id")
	
	// Check if program exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM programs WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if program exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Program not found"})
		return
	}
	
	// Start a transaction to delete program and its routines
	tx, err := db.Begin()
	if err != nil {
		logger.LogError("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	// Delete all routine_exercises for routines in this program
	_, err = tx.Exec(`
		DELETE FROM routine_exercises
		WHERE routine_id IN (SELECT id FROM routines WHERE program_id = $1)
	`, id)
	
	if err != nil {
		tx.Rollback()
		logger.LogError("Failed to delete routine exercises: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete program"})
		return
	}
	
	// Delete all routines in this program
	_, err = tx.Exec("DELETE FROM routines WHERE program_id = $1", id)
	if err != nil {
		tx.Rollback()
		logger.LogError("Failed to delete routines: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete program"})
		return
	}
	
	// Delete the program
	_, err = tx.Exec("DELETE FROM programs WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		logger.LogError("Failed to delete program: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete program"})
		return
	}
	
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		logger.LogError("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Program deleted successfully"})
}

// -------------------- Routine Handlers (Milestone 3) --------------------

func createRoutine(c *gin.Context) {
	var routine Routine
	if err := c.ShouldBindJSON(&routine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validation
	if routine.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	
	if routine.ProgramID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Program ID is required"})
		return
	}
	
	// Check if program exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM programs WHERE id = $1)", routine.ProgramID).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if program exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Program not found"})
		return
	}
	
	// Insert into database
	query := `
		INSERT INTO routines (program_id, name, day_number)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	
	err = db.QueryRow(
		query,
		routine.ProgramID,
		routine.Name,
		routine.DayNumber,
	).Scan(&routine.ID, &routine.CreatedAt, &routine.UpdatedAt)
	
	if err != nil {
		logger.LogError("Failed to create routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create routine"})
		return
	}
	
	c.JSON(http.StatusCreated, routine)
}

func getRoutineByID(c *gin.Context) {
	id := c.Param("id")
	
	var routine Routine
	query := `
		SELECT id, program_id, name, day_number, created_at, updated_at
		FROM routines
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&routine.ID,
		&routine.ProgramID,
		&routine.Name,
		&routine.DayNumber,
		&routine.CreatedAt,
		&routine.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
			return
		}
		
		logger.LogError("Failed to get routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get routine"})
		return
	}
	
	c.JSON(http.StatusOK, routine)
}

func listRoutines(c *gin.Context) {
	limit, offset := getPaginationParams(c)
	programID := c.Query("program_id")
	
	var whereClause string
	var args []interface{}
	
	if programID != "" {
		whereClause = "WHERE program_id = $3"
		args = append(args, programID)
	}
	
	query := fmt.Sprintf(`
		SELECT id, program_id, name, day_number, created_at, updated_at
		FROM routines
		%s
		ORDER BY day_number, name
		LIMIT $1 OFFSET $2
	`, whereClause)
	
	args = append([]interface{}{limit, offset}, args...)
	
	rows, err := db.Query(query, args...)
	if err != nil {
		logger.LogError("Failed to list routines: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list routines"})
		return
	}
	defer rows.Close()
	
	routines := []Routine{}
	for rows.Next() {
		var routine Routine
		if err := rows.Scan(
			&routine.ID,
			&routine.ProgramID,
			&routine.Name,
			&routine.DayNumber,
			&routine.CreatedAt,
			&routine.UpdatedAt,
		); err != nil {
			logger.LogError("Failed to scan routine row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process routines"})
			return
		}
		routines = append(routines, routine)
	}
	
	// Get total count for pagination info
	var total int
	countQuery := "SELECT COUNT(*) FROM routines"
	if programID != "" {
		countQuery += " WHERE program_id = $1"
		err = db.QueryRow(countQuery, programID).Scan(&total)
	} else {
		err = db.QueryRow(countQuery).Scan(&total)
	}
	
	if err != nil {
		logger.LogError("Failed to count routines: %v", err)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": routines,
		"pagination": gin.H{
			"total": total,
			"limit": limit,
			"offset": offset,
		},
	})
}

func updateRoutine(c *gin.Context) {
	id := c.Param("id")
	
	// Check if routine exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM routines WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if routine exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
		return
	}
	
	// Parse request body
	var routine Routine
	if err := c.ShouldBindJSON(&routine); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validation
	if routine.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name is required"})
		return
	}
	
	// Update in database
	query := `
		UPDATE routines
		SET name = $1, day_number = $2, updated_at = NOW()
		WHERE id = $3
		RETURNING id, program_id, name, day_number, created_at, updated_at
	`
	
	err = db.QueryRow(
		query,
		routine.Name,
		routine.DayNumber,
		id,
	).Scan(
		&routine.ID,
		&routine.ProgramID,
		&routine.Name,
		&routine.DayNumber,
		&routine.CreatedAt,
		&routine.UpdatedAt,
	)
	
	if err != nil {
		logger.LogError("Failed to update routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update routine"})
		return
	}
	
	c.JSON(http.StatusOK, routine)
}

func deleteRoutine(c *gin.Context) {
	id := c.Param("id")
	
	// Check if routine exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM routines WHERE id = $1)", id).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if routine exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
		return
	}
	
	// Start a transaction to delete routine and its exercises
	tx, err := db.Begin()
	if err != nil {
		logger.LogError("Failed to begin transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	// Delete all routine_exercises for this routine
	_, err = tx.Exec("DELETE FROM routine_exercises WHERE routine_id = $1", id)
	if err != nil {
		tx.Rollback()
		logger.LogError("Failed to delete routine exercises: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete routine"})
		return
	}
	
	// Delete the routine
	_, err = tx.Exec("DELETE FROM routines WHERE id = $1", id)
	if err != nil {
		tx.Rollback()
		logger.LogError("Failed to delete routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete routine"})
		return
	}
	
	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		logger.LogError("Failed to commit transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Routine deleted successfully"})
}

// -------------------- Routine-Exercise Handlers (Milestone 3) --------------------

func addExerciseToRoutine(c *gin.Context) {
	routineID := c.Param("id")
	
	// Check if routine exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM routines WHERE id = $1)", routineID).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if routine exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
		return
	}
	
	// Parse request body
	var routineExercise RoutineExercise
	if err := c.ShouldBindJSON(&routineExercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Set routine ID from path parameter
	routineExercise.RoutineID, _ = strconv.Atoi(routineID)
	
	// Validation
	if routineExercise.ExerciseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise ID is required"})
		return
	}
	
	// Check if exercise exists
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE id = $1)", routineExercise.ExerciseID).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if exercise exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise not found"})
		return
	}
	
	// Check if the exercise is already in the routine
	err = db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM routine_exercises WHERE routine_id = $1 AND exercise_id = $2)",
		routineExercise.RoutineID,
		routineExercise.ExerciseID,
	).Scan(&exists)
	
	if err != nil {
		logger.LogError("Failed to check if exercise is already in routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise is already in the routine"})
		return
	}
	
	// Insert into database
	query := `
		INSERT INTO routine_exercises (
			routine_id, exercise_id, recommended_sets, recommended_reps, 
			recommended_rpe, recommended_duration, recommended_distance
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	
	err = db.QueryRow(
		query,
		routineExercise.RoutineID,
		routineExercise.ExerciseID,
		routineExercise.RecommendedSets,
		routineExercise.RecommendedReps,
		routineExercise.RecommendedRPE,
		routineExercise.RecommendedDuration,
		routineExercise.RecommendedDistance,
	).Scan(&routineExercise.ID, &routineExercise.CreatedAt, &routineExercise.UpdatedAt)
	
	if err != nil {
		logger.LogError("Failed to add exercise to routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add exercise to routine"})
		return
	}
	
	c.JSON(http.StatusCreated, routineExercise)
}

func getRoutineExercises(c *gin.Context) {
	routineID := c.Param("id")
	
	// Check if routine exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM routines WHERE id = $1)", routineID).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if routine exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Routine not found"})
		return
	}
	
	// Get all exercises in the routine
	query := `
		SELECT 
			re.id, re.routine_id, re.exercise_id, 
			re.recommended_sets, re.recommended_reps, re.recommended_rpe, 
			re.recommended_duration, re.recommended_distance, 
			re.created_at, re.updated_at,
			e.name, e.exercise_type
		FROM routine_exercises re
		JOIN exercises e ON re.exercise_id = e.id
		WHERE re.routine_id = $1
		ORDER BY re.id
	`
	
	rows, err := db.Query(query, routineID)
	if err != nil {
		logger.LogError("Failed to get routine exercises: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get routine exercises"})
		return
	}
	defer rows.Close()
	
	type RoutineExerciseWithDetails struct {
		RoutineExercise
		ExerciseName string `json:"exercise_name"`
		ExerciseType string `json:"exercise_type"`
	}
	
	exercises := []RoutineExerciseWithDetails{}
	for rows.Next() {
		var exercise RoutineExerciseWithDetails
		var exerciseName, exerciseType string
		
		if err := rows.Scan(
			&exercise.ID,
			&exercise.RoutineID,
			&exercise.ExerciseID,
			&exercise.RecommendedSets,
			&exercise.RecommendedReps,
			&exercise.RecommendedRPE,
			&exercise.RecommendedDuration,
			&exercise.RecommendedDistance,
			&exercise.CreatedAt,
			&exercise.UpdatedAt,
			&exerciseName,
			&exerciseType,
		); err != nil {
			logger.LogError("Failed to scan routine exercise row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process routine exercises"})
			return
		}
		
		exercise.ExerciseName = exerciseName
		exercise.ExerciseType = exerciseType
		exercises = append(exercises, exercise)
	}
	
	c.JSON(http.StatusOK, exercises)
}

func updateRoutineExercise(c *gin.Context) {
	routineID := c.Param("id")
	exerciseID := c.Param("exerciseId")
	
	// Check if the routine exercise exists
	var routineExerciseID int
	err := db.QueryRow(
		"SELECT id FROM routine_exercises WHERE routine_id = $1 AND exercise_id = $2",
		routineID,
		exerciseID,
	).Scan(&routineExerciseID)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found in routine"})
			return
		}
		
		logger.LogError("Failed to check if exercise is in routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	// Parse request body
	var routineExercise RoutineExercise
	if err := c.ShouldBindJSON(&routineExercise); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Update in database
	query := `
		UPDATE routine_exercises
		SET 
			recommended_sets = $1, 
			recommended_reps = $2, 
			recommended_rpe = $3, 
			recommended_duration = $4, 
			recommended_distance = $5,
			updated_at = NOW()
		WHERE id = $6
		RETURNING id, routine_id, exercise_id, recommended_sets, recommended_reps, 
			recommended_rpe, recommended_duration, recommended_distance, created_at, updated_at
	`
	
	err = db.QueryRow(
		query,
		routineExercise.RecommendedSets,
		routineExercise.RecommendedReps,
		routineExercise.RecommendedRPE,
		routineExercise.RecommendedDuration,
		routineExercise.RecommendedDistance,
		routineExerciseID,
	).Scan(
		&routineExercise.ID,
		&routineExercise.RoutineID,
		&routineExercise.ExerciseID,
		&routineExercise.RecommendedSets,
		&routineExercise.RecommendedReps,
		&routineExercise.RecommendedRPE,
		&routineExercise.RecommendedDuration,
		&routineExercise.RecommendedDistance,
		&routineExercise.CreatedAt,
		&routineExercise.UpdatedAt,
	)
	
	if err != nil {
		logger.LogError("Failed to update routine exercise: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update routine exercise"})
		return
	}
	
	c.JSON(http.StatusOK, routineExercise)
}

func removeExerciseFromRoutine(c *gin.Context) {
	routineID := c.Param("id")
	exerciseID := c.Param("exerciseId")
	
	// Check if the routine exercise exists
	var exists bool
	err := db.QueryRow(
		"SELECT EXISTS(SELECT 1 FROM routine_exercises WHERE routine_id = $1 AND exercise_id = $2)",
		routineID,
		exerciseID,
	).Scan(&exists)
	
	if err != nil {
		logger.LogError("Failed to check if exercise is in routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Exercise not found in routine"})
		return
	}
	
	// Delete from database
	_, err = db.Exec(
		"DELETE FROM routine_exercises WHERE routine_id = $1 AND exercise_id = $2",
		routineID,
		exerciseID,
	)
	
	if err != nil {
		logger.LogError("Failed to remove exercise from routine: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove exercise from routine"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Exercise removed from routine successfully"})
}

// -------------------- Workout Handlers (Milestone 4) --------------------

func createWorkout(c *gin.Context) {
	var workout Workout
	if err := c.ShouldBindJSON(&workout); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validation
	if workout.UserID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}
	
	// Check if routine exists (if provided)
	if workout.RoutineID > 0 {
		var exists bool
		err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM routines WHERE id = $1)", workout.RoutineID).Scan(&exists)
		if err != nil {
			logger.LogError("Failed to check if routine exists: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			return
		}
		
		if !exists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Routine not found"})
			return
		}
	}
	
	// If no performed_at is provided, use current time
	if workout.PerformedAt.IsZero() {
		workout.PerformedAt = time.Now()
	}
	
	// Insert into database
	query := `
		INSERT INTO workouts (user_id, routine_id, performed_at)
		VALUES ($1, $2, $3)
		RETURNING id, created_at, updated_at
	`
	
	err := db.QueryRow(
		query,
		workout.UserID,
		workout.RoutineID,
		workout.PerformedAt,
	).Scan(&workout.ID, &workout.CreatedAt, &workout.UpdatedAt)
	
	if err != nil {
		logger.LogError("Failed to create workout: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workout"})
		return
	}
	
	// If workout is based on a routine, copy routine exercises to workout sets
	if workout.RoutineID > 0 {
		// Get routine exercises
		rows, err := db.Query(`
			SELECT exercise_id, recommended_sets, recommended_reps, recommended_rpe, 
			recommended_duration, recommended_distance
			FROM routine_exercises
			WHERE routine_id = $1
		`, workout.RoutineID)
		
		if err != nil {
			logger.LogError("Failed to get routine exercises: %v", err)
			// Don't return error, just don't pre-fill the workout
		} else {
			defer rows.Close()
			
			for rows.Next() {
				var exerciseID, recommendedSets, recommendedReps, recommendedDuration int
				var recommendedRPE, recommendedDistance float64
				
				if err := rows.Scan(
					&exerciseID,
					&recommendedSets,
					&recommendedReps,
					&recommendedRPE,
					&recommendedDuration,
					&recommendedDistance,
				); err != nil {
					logger.LogError("Failed to scan routine exercise row: %v", err)
					continue
				}
				
				// Create a workout set for each exercise
				_, err = db.Exec(`
					INSERT INTO workout_sets (
						workout_id, exercise_id, sets, reps, rpe, duration, distance
					)
					VALUES ($1, $2, $3, $4, $5, $6, $7)
				`,
					workout.ID,
					exerciseID,
					recommendedSets,
					recommendedReps,
					recommendedRPE,
					recommendedDuration,
					recommendedDistance,
				)
				
				if err != nil {
					logger.LogError("Failed to create workout set from routine: %v", err)
				}
			}
		}
	}
	
	c.JSON(http.StatusCreated, workout)
}

func getWorkoutByID(c *gin.Context) {
	id := c.Param("id")
	
	var workout Workout
	query := `
		SELECT id, user_id, routine_id, performed_at, created_at, updated_at
		FROM workouts
		WHERE id = $1
	`
	
	err := db.QueryRow(query, id).Scan(
		&workout.ID,
		&workout.UserID,
		&workout.RoutineID,
		&workout.PerformedAt,
		&workout.CreatedAt,
		&workout.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Workout not found"})
			return
		}
		
		logger.LogError("Failed to get workout: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workout"})
		return
	}
	
	// Get routine details if a routine was used
	if workout.RoutineID > 0 {
		var routineName string
		err = db.QueryRow("SELECT name FROM routines WHERE id = $1", workout.RoutineID).Scan(&routineName)
		if err == nil {
			c.JSON(http.StatusOK, gin.H{
				"workout": workout,
				"routine_name": routineName,
			})
			return
		}
	}
	
	c.JSON(http.StatusOK, workout)
}

func listWorkouts(c *gin.Context) {
	limit, offset := getPaginationParams(c)
	userID := c.Query("user_id")
	
	// Validation
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID is required"})
		return
	}
	
	query := `
		SELECT w.id, w.user_id, w.routine_id, w.performed_at, w.created_at, w.updated_at, 
		r.name as routine_name
		FROM workouts w
		LEFT JOIN routines r ON w.routine_id = r.id
		WHERE w.user_id = $1
		ORDER BY w.performed_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := db.Query(query, userID, limit, offset)
	if err != nil {
		logger.LogError("Failed to list workouts: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list workouts"})
		return
	}
	defer rows.Close()
	
	type WorkoutWithRoutineName struct {
		Workout
		RoutineName string `json:"routine_name,omitempty"`
	}
	
	workouts := []WorkoutWithRoutineName{}
	for rows.Next() {
		var workout WorkoutWithRoutineName
		var routineName sql.NullString
		
		if err := rows.Scan(
			&workout.ID,
			&workout.UserID,
			&workout.RoutineID,
			&workout.PerformedAt,
			&workout.CreatedAt,
			&workout.UpdatedAt,
			&routineName,
		); err != nil {
			logger.LogError("Failed to scan workout row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process workouts"})
			return
		}
		
		if routineName.Valid {
			workout.RoutineName = routineName.String
		}
		
		workouts = append(workouts, workout)
	}
	
	// Get total count for pagination info
	var total int
	countQuery := "SELECT COUNT(*) FROM workouts WHERE user_id = $1"
	err = db.QueryRow(countQuery, userID).Scan(&total)
	
	if err != nil {
		logger.LogError("Failed to count workouts: %v", err)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": workouts,
		"pagination": gin.H{
			"total": total,
			"limit": limit,
			"offset": offset,
		},
	})
}

// -------------------- Workout Set Handlers (Milestone 4) --------------------

func addWorkoutSet(c *gin.Context) {
	workoutID := c.Param("id")
	
	// Check if workout exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM workouts WHERE id = $1)", workoutID).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if workout exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workout not found"})
		return
	}
	
	// Parse request body
	var workoutSet WorkoutSet
	if err := c.ShouldBindJSON(&workoutSet); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Set workout ID from path parameter
	workoutSet.WorkoutID, _ = strconv.Atoi(workoutID)
	
	// Validation
	if workoutSet.ExerciseID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise ID is required"})
		return
	}
	
	// Check if exercise exists
	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM exercises WHERE id = $1)", workoutSet.ExerciseID).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if exercise exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Exercise not found"})
		return
	}
	
	// Get exercise type for validation
	var exerciseType string
	err = db.QueryRow("SELECT exercise_type FROM exercises WHERE id = $1", workoutSet.ExerciseID).Scan(&exerciseType)
	if err != nil {
		logger.LogError("Failed to get exercise type: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	// Validate based on exercise type
	if exerciseType == "weight_reps" {
		if workoutSet.Sets <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Sets must be greater than 0 for weight_reps exercise"})
			return
		}
		if workoutSet.Reps <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Reps must be greater than 0 for weight_reps exercise"})
			return
		}
	} else if exerciseType == "duration_only" {
		if workoutSet.Duration <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duration must be greater than 0 for duration_only exercise"})
			return
		}
	} else if exerciseType == "distance_time" {
		if workoutSet.Distance <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Distance must be greater than 0 for distance_time exercise"})
			return
		}
		if workoutSet.Duration <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Duration must be greater than 0 for distance_time exercise"})
			return
		}
	}
	
	// Insert into database
	query := `
		INSERT INTO workout_sets (workout_id, exercise_id, sets, reps, weight, rpe, duration, distance)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	
	err = db.QueryRow(
		query,
		workoutSet.WorkoutID,
		workoutSet.ExerciseID,
		workoutSet.Sets,
		workoutSet.Reps,
		workoutSet.Weight,
		workoutSet.RPE,
		workoutSet.Duration,
		workoutSet.Distance,
	).Scan(&workoutSet.ID, &workoutSet.CreatedAt, &workoutSet.UpdatedAt)
	
	if err != nil {
		logger.LogError("Failed to add workout set: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add workout set"})
		return
	}
	
	c.JSON(http.StatusCreated, workoutSet)
}

func getWorkoutSets(c *gin.Context) {
	workoutID := c.Param("id")
	
	// Check if workout exists
	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM workouts WHERE id = $1)", workoutID).Scan(&exists)
	if err != nil {
		logger.LogError("Failed to check if workout exists: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}
	
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "Workout not found"})
		return
	}
	
	// Get all sets in the workout
	query := `
		SELECT 
			ws.id, ws.workout_id, ws.exercise_id, 
			ws.sets, ws.reps, ws.weight, ws.rpe, ws.duration, ws.distance, 
			ws.created_at, ws.updated_at,
			e.name as exercise_name, e.exercise_type
		FROM workout_sets ws
		JOIN exercises e ON ws.exercise_id = e.id
		WHERE ws.workout_id = $1
		ORDER BY ws.id
	`
	
	rows, err := db.Query(query, workoutID)
	if err != nil {
		logger.LogError("Failed to get workout sets: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get workout sets"})
		return
	}
	defer rows.Close()
	
	type WorkoutSetWithDetails struct {
		WorkoutSet
		ExerciseName string `json:"exercise_name"`
		ExerciseType string `json:"exercise_type"`
	}
	
	sets := []WorkoutSetWithDetails{}
	for rows.Next() {
		var set WorkoutSetWithDetails
		var exerciseName, exerciseType string
		
		if err := rows.Scan(
			&set.ID,
			&set.WorkoutID,
			&set.ExerciseID,
			&set.Sets,
			&set.Reps,
			&set.Weight,
			&set.RPE,
			&set.Duration,
			&set.Distance,
			&set.CreatedAt,
			&set.UpdatedAt,
			&exerciseName,
			&exerciseType,
		); err != nil {
			logger.LogError("Failed to scan workout set row: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process workout sets"})
			return
		}
		
		set.ExerciseName = exerciseName
		set.ExerciseType = exerciseType
		sets = append(sets, set)
	}
	
	c.JSON(http.StatusOK, sets)
}

func updateWorkoutSet(c *gin.Context) {
	workoutID := c.Param("id")
	setID := c.Param("setId")
	
	// Check if the workout set exists