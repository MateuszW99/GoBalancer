var builder = WebApplication.CreateBuilder(args);

// Add services to the container.

var app = builder.Build();

// Configure the HTTP request pipeline.
if (app.Environment.IsDevelopment())
{
}

app.UseHttpsRedirection();

app.MapGet("/healthcheck", () => new HealthCheckResponse("1.0.0", DateTime.UtcNow));

app.Run();

record HealthCheckResponse(string Version, DateTime CurrentDate);
