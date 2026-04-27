export type Role = 'user' | 'admin'

export type Gender = 'male' | 'female'

export type DocumentType = 'passport'

// --- Auth ---

export interface SignUpRequest {
  email: string
  password: string
  phone_number: string
}

export interface SignUpResponse {
  user_id: number
}

export interface SignInRequest {
  email: string
  password: string
}

export interface SignInResponse {
  user_id: number
  token: string
}

// --- User ---

export interface UserResponse {
  id: number
  email: string
  phone_number: string
  role: Role
  /** ISO 8601 UTC */
  created_at: string
}

// --- Passengers ---

/** Full passenger profile — for POST /users/:id/passengers */
export interface PassengerInput {
  first_name: string
  last_name: string
  /** Optional */
  middle_name?: string
  /** ISO 8601 date: YYYY-MM-DD */
  birth_date: string
  gender: Gender
  document_number: string
  document_type: DocumentType
  /** ISO 3166-1 alpha-3, e.g. "RUS" */
  citizenship: string
}

export interface PassengerResponse {
  id: number
  user_id: number
  first_name: string
  last_name: string
  middle_name?: string
  /** ISO 8601 date: YYYY-MM-DD */
  birth_date: string
  gender: Gender
  document_number: string
  document_type: DocumentType
  citizenship: string
}

export interface CreatePassengerResponse {
  passenger_id: number
}
