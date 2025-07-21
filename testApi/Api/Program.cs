var builder = WebApplication.CreateBuilder(args);

// Add services to the container.
builder.Services.AddHttpLogging(_ => {});
var app = builder.Build();

// Configure the HTTP request pipeline.
app.UseHttpLogging();

if (app.Environment.IsDevelopment())
{
}

app.UseHttpsRedirection();

app.MapGet("/healthcheck", () => new HealthCheckResponse("1.0.0", DateTime.UtcNow));

app.MapGet("/hello-world", () => new { Message = "Hello, World!" });

app.Run();

record HealthCheckResponse(string Version, DateTime CurrentDate);
