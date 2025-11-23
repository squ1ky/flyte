package ru.kpfu.flyte.flight_service.dto.aircraft;

import lombok.AllArgsConstructor;
import lombok.Getter;
import lombok.NoArgsConstructor;

@Getter
@NoArgsConstructor
@AllArgsConstructor
public class AircraftResponseDto {
    private Long id;
    private String code;
    private String name;
    private int totalSeats;
}
