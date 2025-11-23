package ru.kpfu.flyte.flight_service.dto.seat;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;

@Getter
@NoArgsConstructor
@AllArgsConstructor
public class SeatDto {
    private Long id;
    private int rowNumber;
    private String seatColumn;
    private String seatNumber;
    private String cabinClass;
    private boolean window;
    private boolean aisle;
    private boolean exitRow;
}
