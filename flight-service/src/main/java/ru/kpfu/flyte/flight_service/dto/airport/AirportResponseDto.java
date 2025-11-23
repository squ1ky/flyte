package ru.kpfu.flyte.flight_service.dto.airport;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Getter;
import lombok.NoArgsConstructor;

@Getter
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class AirportResponseDto {
    private Long id;
    private String code;
    private String name;
    private String city;
    private String country;
}
