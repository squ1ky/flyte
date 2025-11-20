package ru.kpfu.flyte.flight_service.model;

import jakarta.persistence.*;
import lombok.Getter;
import lombok.NoArgsConstructor;
import lombok.Setter;

import java.time.LocalDateTime;

@Entity
@Table(
        name = "aircraft_seats",
        uniqueConstraints = {
                @UniqueConstraint(
                        name = "uq_aircraft_seat",
                        columnNames = {"aircraft_id", "seat_number"}
                )
        }
)
@NoArgsConstructor
@Getter
@Setter
public class AircraftSeat {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY, optional = false)
    @JoinColumn(name = "aircraft_id", nullable = false)
    private Aircraft aircraft;

    @Column(name = "row_number", nullable = false)
    private Integer rowNumber;

    @Column(name = "seat_column", nullable = false, length = 2)
    private String seatColumn;

    @Column(name = "seat_number", nullable = false, length = 10)
    private String seatNumber;

    @Column(name = "cabin_class", nullable = false, length = 50)
    private String cabinClass;

    @Column(name = "is_window", nullable = false)
    private boolean window;

    @Column(name = "is_aisle", nullable = false)
    private boolean aisle;

    @Column(name = "is_exit_row", nullable = false)
    private boolean exitRow;

    @Column(name = "created_at", nullable = false)
    private LocalDateTime createdAt;

    @Column(name = "updated_at", nullable = false)
    private LocalDateTime updatedAt;

    @PrePersist
    void onCreate() {
        LocalDateTime now = LocalDateTime.now();
        this.createdAt = now;
        this.updatedAt = now;
    }

    @PreUpdate
    void onUpdate() {
        this.updatedAt = LocalDateTime.now();
    }
}
