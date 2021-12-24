attended := map[string]bool{
    "Ann": true,
    "Joe": true,
}

if attended[person] { // will be false if person is not in the map
    fmt.Println(person, "was at the meeting")
}
