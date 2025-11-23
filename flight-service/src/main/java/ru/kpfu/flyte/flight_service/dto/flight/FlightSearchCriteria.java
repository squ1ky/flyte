package ru.kpfu.flyte.flight_service.dto.flight;

import lombok.Getter;
import lombok.Setter;

import java.time.LocalDateTime;

@Getter
@Setter
public class FlightSearchCriteria {
    private String originAirportCode;
    private String destinationAirportCode;
    private LocalDateTime departureDate;
}
